# -*- coding: utf-8 -*-
"""
从 Go 二进制中提取 //go:embed 的静态资源。

原理：Go 把 embed 的文件表编译成连续的结构体数组：
    struct file { name string; data string; hash [16]byte }
即 48 字节：name.ptr(8) name.len(8) data.ptr(8) data.len(8) hash(16)

文件名字符串本身存在 .rodata 里，其虚拟地址就是 name.ptr。
所以只要：
  1) 在文件里搜到 "assets/xxx.js" 的字节，
  2) 算出它的虚拟地址 VA，
  3) 在二进制里搜 (VA, len) 这 16 字节小端序列，
  4) 命中处 +16 就是 data.ptr / data.len，
就能精确取出该文件的原始内容（长度来自 data.len，不靠猜）。
"""
import struct
import sys
import os

ELF = sys.argv[1] if len(sys.argv) > 1 else "full.bin"
OUT = sys.argv[2] if len(sys.argv) > 2 else "extracted"

data = open(ELF, "rb").read()
print("载入 %s, %.1f MB" % (ELF, len(data) / 1048576))

# ---------- 解析 ELF64 program headers，建立 VA <-> 文件偏移映射 ----------
assert data[:4] == b"\x7fELF", "不是 ELF"
assert data[4] == 2, "非 64 位"
assert data[5] == 1, "非小端"

e_phoff = struct.unpack_from("<Q", data, 0x20)[0]
e_phentsize = struct.unpack_from("<H", data, 0x36)[0]
e_phnum = struct.unpack_from("<H", data, 0x38)[0]

segments = []  # (vaddr, filesz, offset)
for i in range(e_phnum):
    base = e_phoff + i * e_phentsize
    p_type = struct.unpack_from("<I", data, base)[0]
    if p_type != 1:  # PT_LOAD
        continue
    p_offset = struct.unpack_from("<Q", data, base + 8)[0]
    p_vaddr = struct.unpack_from("<Q", data, base + 16)[0]
    p_filesz = struct.unpack_from("<Q", data, base + 32)[0]
    segments.append((p_vaddr, p_filesz, p_offset))

print("PT_LOAD 段数量:", len(segments))


def off_to_va(off):
    for vaddr, filesz, p_off in segments:
        if p_off <= off < p_off + filesz:
            return vaddr + (off - p_off)
    return None


def va_to_off(va):
    for vaddr, filesz, p_off in segments:
        if vaddr <= va < vaddr + filesz:
            return p_off + (va - vaddr)
    return None


# ---------- 找出所有 assets/ 开头的文件名字符串 ----------
names = {}
for m in __import__("re").finditer(
    rb'assets/[^\x00-\x1f/\\:*?"<>|\x7f]{1,80}\.(?:js|css|png|svg|jpg|jpeg|gif|woff2?|ttf|ico)',
    data,
):
    raw = m.group(0)
    try:
        name = raw.decode("utf-8")
    except UnicodeDecodeError:
        continue
    off = m.start()
    va = off_to_va(off)
    if va is None:
        continue
    names.setdefault(name, (off, va, len(raw)))  # 注意：len 必须是"字节长度"，中文名一个字 3 字节

print("发现 assets 文件名:", len(names))

# ---------- 对每个文件名定位 file 结构并提取内容 ----------
os.makedirs(OUT, exist_ok=True)
ok = 0
fail = []
seen_ranges = []

for name, (name_off, name_va, name_len) in sorted(names.items()):
    key = struct.pack("<QQ", name_va, name_len)
    # 结构体里 name 字段紧邻 data 字段，搜索 (name.ptr,name.len) 序列
    pos = None
    start = 0
    while True:
        p = data.find(key, start)
        if p < 0:
            break
        # 命中点之后必须有 16 字节（data.ptr + data.len）
        if p + 32 > len(data):
            break
        d_ptr, d_len = struct.unpack_from("<QQ", data, p + 16)
        # 长度合理性校验：非 0、不超过二进制大小
        if 0 < d_len < len(data):
            target = va_to_off(d_ptr)
            if target is not None and target + d_len <= len(data):
                pos = (target, d_len)
                break
        start = p + 1
    if pos is None:
        fail.append(name)
        continue
    target, d_len = pos
    content = data[target:target + d_len]
    path = os.path.join(OUT, name)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "wb") as f:
        f.write(content)
    ok += 1
    seen_ranges.append((target, target + d_len, name))

print("成功提取:", ok, " 失败:", len(fail))
if fail:
    print("失败列表(前10):", fail[:10])

# 统计
total = 0
for root, _, files in os.walk(OUT):
    for f in files:
        total += os.path.getsize(os.path.join(root, f))
print("提取总大小: %.1f MB" % (total / 1048576))
