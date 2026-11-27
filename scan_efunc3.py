# -*- coding: utf-8 -*-
"""
v3: 生成 EFunc/utils 待实现函数清单（按族分组 + 参数个数 + 典型调用）
"""
import os, re, collections, json

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "ky")
OUT = os.path.join(os.path.dirname(ROOT), "EFunc待实现清单.md")
ID = r'[A-Za-z_\u4e00-\u9fff][\w\u4e00-\u9fff]*'

# 已确认是本地定义的误报
BLACKLIST = {"GetList2", "KaClassUp带调价信息", "ToId", "D订单支付回调",
             "Q取通道名称", "Update", "Find", "First", "Count"}

# 已由 app/utils 本地实现（无需补）
ALREADY = {"Z正则_校验密码", "Z正则_校验代理用户名", "Z正则_校验email", "Z正则_校验用户名",
           "Z正则_校验纯数字", "Z正则_校验手机号", "Z正则_是否英数", "X系统_权限检测"}

FAMILY = [
    ("泛型/语法糖", re.compile(r'^S三元$|^Map_键名是否存在$')),
    ("文本处理 W文本_*", re.compile(r'^W文本_')),
    ("文件 IO W文件_*", re.compile(r'^W文件_')),
    ("数组/切片 S数组_*", re.compile(r'^S数组_')),
    ("浮点运算 Float64*", re.compile(r'^(Float64|Int64到Float64|Int32ToBytes)')),
    ("类型转换 D到*", re.compile(r'^D到')),
    ("编码 B编码_*", re.compile(r'^B编码_')),
    ("字节集 Z字节集_*", re.compile(r'^Z字节集_')),
    ("时间 S时间_*", re.compile(r'^S时间_')),
    ("正则 Z正则_*", re.compile(r'^Z正则_')),
    ("校验 J校验_*", re.compile(r'^J校验_')),
    ("进程/目录 C/M 前缀", re.compile(r'^[CM](程序|目录)_')),
    ("图片 T图片_*", re.compile(r'^T图片_')),
    ("底层/汇编 H汇编_*", re.compile(r'^H汇编_')),
]


def read(p):
    with open(p, encoding="utf-8", errors="replace") as fh:
        return fh.read()


def strip_literals(src):
    out, i, n = [], 0, len(src)
    while i < n:
        c = src[i]
        if c == '"':
            out.append('""'); i += 1
            while i < n:
                if src[i] == '\\': i += 2; continue
                if src[i] == '"': i += 1; break
                i += 1
        elif c == '`':
            out.append('``'); i += 1
            while i < n and src[i] != '`': i += 1
            i += 1
        elif c == "'":
            out.append("''"); i += 1
            while i < n:
                if src[i] == '\\': i += 2; continue
                if src[i] == "'": i += 1; break
                i += 1
        elif c == '/' and i + 1 < n and src[i + 1] == '/':
            while i < n and src[i] != '\n': i += 1
        elif c == '/' and i + 1 < n and src[i + 1] == '*':
            i += 2
            while i + 1 < n and not (src[i] == '*' and src[i + 1] == '/'): i += 1
            i += 2
        else:
            out.append(c); i += 1
    return ''.join(out)


def split_args(s):
    """按顶层逗号切分参数"""
    parts, depth, cur, instr = [], 0, [], None
    i = 0
    while i < len(s):
        c = s[i]
        if instr:
            if c == '\\': cur.append(s[i:i+2]); i += 2; continue
            if c == instr: instr = None
            cur.append(c)
        elif c in '"`\'':
            instr = c; cur.append(c)
        elif c in '([{':
            depth += 1; cur.append(c)
        elif c in ')]}':
            depth -= 1; cur.append(c)
        elif c == ',' and depth == 0:
            parts.append(''.join(cur).strip()); cur = []
        else:
            cur.append(c)
        i += 1
    tail = ''.join(cur).strip()
    if tail or parts:
        parts.append(tail)
    return parts


go_files = []
for dp, dn, fn in os.walk(ROOT):
    if ".git" in dp.split(os.sep):
        continue
    for f in fn:
        if f.endswith(".go"):
            go_files.append(os.path.join(dp, f))

# 逐文件确定 EFunc 别名
file_alias = {}
for p in go_files:
    raw = read(p)
    if '"EFunc/utils"' not in raw:
        continue
    # 只认 import 块内、显式指向 EFunc/utils 的行
    for line in raw.splitlines():
        if '"EFunc/utils"' in line and 'import' not in line:
            m = re.match(r'\s*(' + ID + r'|\.)\s+"EFunc/utils"', line)
            file_alias[p] = m.group(1) if m else "utils"
            break

# ---------- 点导入文件的噪声过滤素材 ----------
GO_KEYWORDS = {
    "If", "For", "Switch", "Select", "Return", "Func", "Go", "Defer", "Range",
    "Case", "Default", "Else", "Chan", "Interface", "Struct", "Map", "Make",
    "New", "Len", "Cap", "Append", "Copy", "Delete", "Close", "Panic", "Recover",
    "Print", "Println", "String", "Int", "Int8", "Int16", "Int32", "Int64",
    "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Uintptr", "Float32",
    "Float64", "Bool", "Byte", "Rune", "Error", "Any", "Nil", "True", "False",
    "Iota", "Var", "Const", "Type", "Package", "Import",
}
GORM_NOISE = {"Where", "Select", "Order", "Group", "Count", "First", "Find",
              "Updates", "Clauses", "Sum", "Pluck", "Joins", "Preload", "Model",
              "Table", "Limit", "Offset", "Delete", "Create", "Save", "Raw", "Exec"}

pkg_defs = collections.defaultdict(set)
for p in go_files:
    code = strip_literals(read(p))
    d = os.path.dirname(p)
    for m in re.finditer(r'^func\s+(?:\([^)]*\)\s*)?(' + ID + r')\s*[\(\[]', code, re.M):
        pkg_defs[d].add(m.group(1))
    for m in re.finditer(r'^type\s+(' + ID + r')', code, re.M):
        pkg_defs[d].add(m.group(1))
    for m in re.finditer(r'^(?:var|const)\s+(' + ID + r')', code, re.M):
        pkg_defs[d].add(m.group(1))
    for blk in re.finditer(r'^(?:var|const)\s*\((.*?)^\)', code, re.M | re.S):
        for m in re.finditer(r'^\t*(' + ID + r')\s', blk.group(1), re.M):
            pkg_defs[d].add(m.group(1))

response_exports = set()
MDL = os.path.join(ROOT, "app", "models", "response")
if os.path.isdir(MDL):
    for f in os.listdir(MDL):
        if f.endswith(".go"):
            code = strip_literals(read(os.path.join(MDL, f)))
            for m in re.finditer(r'^func\s+(?:\([^)]*\)\s*)?([A-Z]' + ID[1:] + r')\s*[\(\[]', code, re.M):
                response_exports.add(m.group(1))
            for m in re.finditer(r'^type\s+([A-Z]' + ID[1:] + r')', code, re.M):
                response_exports.add(m.group(1))
            for blk in re.finditer(r'^type\s*\((.*?)^\)', code, re.M | re.S):
                for m in re.finditer(r'^\t*([A-Z]' + ID[1:] + r')\s', blk.group(1), re.M):
                    response_exports.add(m.group(1))

syms = collections.Counter()
sym_files = collections.defaultdict(set)
sym_argc = collections.defaultdict(collections.Counter)
sym_sample = {}

def extract_args(raw, open_idx):
    """从 raw[open_idx]=='(' 起配对括号，跳过字符串/注释，返回 (参数文本, 结束位置)"""
    depth, i, n = 1, open_idx + 1, len(raw)
    instr = None
    while i < n:
        c = raw[i]
        if instr:
            if c == '\\': i += 2; continue
            if c == instr: instr = None
            i += 1
        elif c in '"`\'':
            instr = c; i += 1
        elif c == '/' and i + 1 < n and raw[i + 1] == '/':
            while i < n and raw[i] != '\n': i += 1
        elif c == '(':
            depth += 1; i += 1
        elif c == ')':
            depth -= 1
            if depth == 0:
                return raw[open_idx + 1:i], i
            i += 1
        else:
            i += 1
    return raw[open_idx + 1:], n


for p, alias in file_alias.items():
    raw = read(p)
    code = strip_literals(raw)
    rel = os.path.relpath(p, ROOT)
    pdir = os.path.dirname(p)

    if alias == ".":
        # 点导入：裸调用需要剔除本包/本地变量/其它点导入包
        local = set(re.findall(r'(?<![\w.])(' + ID + r')\s*:=', code))
        for m in re.finditer(r'(?<![\w.])((?:' + ID + r'\s*,\s*)+' + ID + r')\s*:=', code):
            for part in m.group(1).split(','):
                local.add(part.strip())
        local |= set(re.findall(r'\bvar\s+(' + ID + r')', code))
        pat = r'(?<![\w.])(' + ID + r')(?=\s*\()'
    else:
        local = set()
        pat = r'(?<![\w.])' + re.escape(alias) + r'\.(' + ID + r')'

    for m in re.finditer(pat, code):
        name = m.group(1)
        if name in BLACKLIST or 'A' > name[0] or name[0] > 'Z':
            continue
        if alias == ".":
            if name in GO_KEYWORDS or name in local or name in GORM_NOISE:
                continue
            if name in pkg_defs[pdir] or name in response_exports:
                continue
        syms[name] += 1
        sym_files[name].add(rel)

        if name in sym_sample and syms[name] > 40:
            continue
        # 在原文 raw 上重新定位调用点，保证偏移正确
        if alias == ".":
            cand = re.search(r'(?<![\w.])' + re.escape(name) + r'\s*\(', raw)
        else:
            cand = re.search(r'(?<![\w.])' + re.escape(alias) + r'\.' + re.escape(name) + r'\s*\(', raw)
        if not cand:
            continue
        open_idx = raw.index('(', cand.start())
        args_text, end_idx = extract_args(raw, open_idx)
        argv = split_args(args_text)
        if argv and argv != ['']:
            sym_argc[name][len(argv)] += 1
        else:
            sym_argc[name][0] += 1
        if name not in sym_sample:
            line_start = raw.rfind('\n', 0, cand.start()) + 1
            sym_sample[name] = raw[line_start:end_idx + 1].strip().replace('\n', ' ')

# 分组
groups = collections.OrderedDict((k, []) for k, _ in FAMILY)
groups["其它"] = []
for name in syms:
    for gname, rx in FAMILY:
        if rx.search(name):
            groups[gname].append(name)
            break
    else:
        groups["其它"].append(name)

lines = []
lines.append("# KuaiYanGo2 → EFunc/utils 移植清单")
lines.append("")
lines.append("> 由静态扫描生成。共 **%d** 个外部符号，分布在 **%d** 个文件中，总引用 **%d** 处。"
             % (len(syms), len(set().union(*sym_files.values())), sum(syms.values())))
lines.append("")
lines.append("| 族 | 符号数 | 引用数 | 难度 |")
lines.append("|---|---|---|---|")
DIFF = {"泛型/语法糖": "低", "文本处理 W文本_*": "低", "文件 IO W文件_*": "低",
        "数组/切片 S数组_*": "低", "浮点运算 Float64*": "中", "类型转换 D到*": "低",
        "编码 B编码_*": "低", "字节集 Z字节集_*": "低", "时间 S时间_*": "低",
        "正则 Z正则_*": "已有", "校验 J校验_*": "低", "进程/目录 C/M 前缀": "低",
        "图片 T图片_*": "中", "底层/汇编 H汇编_*": "高", "其它": "?"}
for g, names in groups.items():
    if not names:
        continue
    names.sort(key=lambda x: -syms[x])
    lines.append("| %s | %d | %d | %s |" % (g, len(names), sum(syms[n] for n in names), DIFF.get(g, "?")))
lines.append("")
lines.append("---")
lines.append("")
for g, names in groups.items():
    if not names:
        continue
    lines.append("## %s（%d 个）" % (g, len(names)))
    lines.append("")
    lines.append("| 符号 | 引用 | 文件 | 参数个数 | 状态 | 典型调用 |")
    lines.append("|---|---|---|---|---|---|")
    for n in names:
        arc = "/".join(str(a) for a, _ in sym_argc[n].most_common()) or "-"
        state = "✅ 已有" if n in ALREADY else ""
        sample = sym_sample.get(n, "")[:90].replace('|', '\\|')
        lines.append("| `%s` | %d | %d | %s | %s | `%s` |" % (n, syms[n], len(sym_files[n]), arc, state, sample))
    lines.append("")

with open(OUT, "w", encoding="utf-8") as fh:
    fh.write("\n".join(lines))

print("符号总数: %d  总引用: %d  涉及文件: %d"
      % (len(syms), sum(syms.values()), len(set().union(*sym_files.values()))))
print("-" * 60)
for g, names in groups.items():
    if names:
        print("%-22s %3d 个   %4d 次" % (g, len(names), sum(syms[n] for n in names)))
print("-" * 60)
print("已在 app/utils 本地实现:", len([n for n in syms if n in ALREADY]))
print("报告 ->", OUT)
