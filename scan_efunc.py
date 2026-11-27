# -*- coding: utf-8 -*-
"""扫描 KuaiYanGo2 中对 EFunc/utils 的依赖，量化移植工作量。"""
import os, re, collections, json

ROOT = r"C:\Users\Administrator\WorkBuddy\2026-11-27-09-43-34\ky"
ID = r'[A-Za-z_\u4e00-\u9fff][\w\u4e00-\u9fff]*'

GO_KEYWORDS = {
    "if", "for", "switch", "select", "return", "func", "go", "defer", "range",
    "case", "default", "else", "chan", "interface", "struct", "map", "make",
    "new", "len", "cap", "append", "copy", "delete", "close", "panic", "recover",
    "print", "println", "complex", "real", "imag", "string", "int", "int8",
    "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64",
    "uintptr", "float32", "float64", "bool", "byte", "rune", "error", "any",
    "nil", "true", "false", "iota", "var", "const", "type", "package", "import",
}

go_files = []
for dp, dn, fn in os.walk(ROOT):
    if ".git" in dp.split(os.sep):
        continue
    for f in fn:
        if f.endswith(".go"):
            go_files.append(os.path.join(dp, f))


def read(p):
    with open(p, encoding="utf-8", errors="replace") as fh:
        return fh.read()


# ---------- 1. 收集每个 package(目录) 内定义的顶层名字 ----------
pkg_defs = collections.defaultdict(set)
pkg_funcs = collections.defaultdict(set)
for p in go_files:
    src = read(p)
    d = os.path.dirname(p)
    for m in re.finditer(r'^func\s+(?:\([^)]*\)\s*)?(' + ID + r')\s*[\(\[]', src, re.M):
        pkg_defs[d].add(m.group(1))
        pkg_funcs[d].add(m.group(1))
    for m in re.finditer(r'^type\s+(' + ID + r')', src, re.M):
        pkg_defs[d].add(m.group(1))
    for m in re.finditer(r'^(?:var|const)\s+(' + ID + r')', src, re.M):
        pkg_defs[d].add(m.group(1))
    # var/const (...) 块
    for blk in re.finditer(r'^(?:var|const)\s*\((.*?)^\)', src, re.M | re.S):
        for m in re.finditer(r'^\t*(' + ID + r')[\s=]', blk.group(1), re.M):
            pkg_defs[d].add(m.group(1))

# ---------- 2. 其它点导入包的导出名（本仓库内可解析的） ----------
MDL = os.path.join(ROOT, "app", "models", "response")
response_exports = set()
if os.path.isdir(MDL):
    for f in os.listdir(MDL):
        if f.endswith(".go"):
            src = read(os.path.join(MDL, f))
            for m in re.finditer(r'^func\s+(?:\([^)]*\)\s*)?([A-Z]' + ID[1:] + r')\s*[\(\[]', src, re.M):
                response_exports.add(m.group(1))
            for m in re.finditer(r'^type\s+([A-Z]' + ID[1:] + r')', src, re.M):
                response_exports.add(m.group(1))

# ---------- 3. 逐文件提取 EFunc 符号 ----------
ALIAS_MAP = {"utils", "utils2"}
syms = collections.Counter()        # 符号 -> 引用次数
sym_files = collections.defaultdict(set)  # 符号 -> 文件集合
sym_kind = {}                       # 符号 -> 判定类型
dot_files = []
alias_files = []
weird = collections.Counter()       # 点导入里无法判定归属的名字

for p in go_files:
    src = read(p)
    if '"EFunc/utils"' not in src:
        continue
    rel = os.path.relpath(p, ROOT)
    alias = None
    for line in src.splitlines():
        if '"EFunc/utils"' in line:
            m = re.match(r'\s*(' + ID + r'|\.)\s+"EFunc/utils"', line)
            alias = m.group(1) if m else "utils"
            break

    called = {}
    if alias == ".":
        dot_files.append(rel)
        for m in re.finditer(r'(?<![\w.])(' + ID + r')\s*\(', src):
            name = m.group(1)
            if name in GO_KEYWORDS:
                continue
            if name in pkg_funcs[os.path.dirname(p)]:
                continue
            if name in response_exports:
                continue
            # 排除本文件内直接 func 定义（避免误算）
            called[name] = called.get(name, 0) + 1
        # 非调用形式（类型/常量/变量）
        for m in re.finditer(r'(?<![\w.])(' + ID + r')\s*\{', src):
            name = m.group(1)
            if name in GO_KEYWORDS or name in pkg_defs[os.path.dirname(p)]:
                continue
            if name in response_exports:
                continue
            if name not in called:
                called[name] = called.get(name, 0) + 1
                sym_kind.setdefault(name, "type/var?")
    else:
        alias_files.append(rel)
        for m in re.finditer(r'(?<![\w.])' + re.escape(alias) + r'\.(' + ID + r')', src):
            name = m.group(1)
            called[name] = called.get(name, 0) + 1
        for name in called:
            # 判定：后面紧跟 ( 则为函数
            if re.search(re.escape(alias) + r'\.' + re.escape(name) + r'\s*\(', src):
                sym_kind.setdefault(name, "func")
            else:
                sym_kind.setdefault(name, "type/var/const")

    for name, n in called.items():
        syms[name] += n
        sym_files[name].add(rel)

# ---------- 4. 输出 ----------
print("=" * 78)
print("扫描范围: %d 个 .go 文件, 其中 %d 个引用 EFunc/utils" % (len(go_files), len(dot_files) + len(alias_files)))
print("点导入(裸调用)文件: %d | 别名导入文件: %d" % (len(dot_files), len(alias_files)))
print("识别到的 EFunc 符号总数: %d" % len(syms))
print("=" * 78)

print("\n【按引用次数排序】符号 | 引用次数 | 涉及文件数 | 判定类型")
for name, n in syms.most_common():
    print("%-42s %5d  %4d  %s" % (name, n, len(sym_files[name]), sym_kind.get(name, "func?")))

with open(os.path.join(os.path.dirname(ROOT), "efunc_syms.json"), "w", encoding="utf-8") as fh:
    json.dump(
        {
            "symbols": [
                {"name": k, "calls": v, "files": sorted(sym_files[k]), "kind": sym_kind.get(k, "func?")}
                for k, v in syms.most_common()
            ],
            "dot_files": dot_files,
            "alias_files": alias_files,
        },
        fh,
        ensure_ascii=False,
        indent=2,
    )
print("\n明细已写入 efunc_syms.json")
