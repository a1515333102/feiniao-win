# -*- coding: utf-8 -*-
"""
v2: 精确提取 KuaiYanGo2 对 EFunc/utils 的依赖。
关键改进：
1. 剥离字符串字面量/注释，消除 SQL 关键字噪声（SUM/Where/CONCAT...）
2. 剔除本文件局部声明（:=、var、for range）
3. 利用 Go 导出规则：跨包可见的标识符首字符必须是大写 ASCII 字母
   -> 纯中文开头、小写开头的名字不可能是 EFunc 导出的，直接排除
4. 交叉验证：在 `utils.X` 别名导入文件中出现的符号 = 高置信
"""
import os, re, collections, json

ROOT = r"C:\Users\Administrator\WorkBuddy\2026-11-27-09-43-34\ky"
ID = r'[A-Za-z_\u4e00-\u9fff][\w\u4e00-\u9fff]*'

GO_KEYWORDS = {
    "If", "For", "Switch", "Select", "Return", "Func", "Go", "Defer", "Range",
    "Case", "Default", "Else", "Chan", "Interface", "Struct", "Map", "Make",
    "New", "Len", "Cap", "Append", "Copy", "Delete", "Close", "Panic", "Recover",
    "Print", "Println", "String", "Int", "Int8", "Int16", "Int32", "Int64",
    "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Uintptr", "Float32",
    "Float64", "Bool", "Byte", "Rune", "Error", "Any", "Nil", "True", "False",
    "Iota", "Var", "Const", "Type", "Package", "Import",
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


def strip_literals(src):
    """去掉字符串、rune、注释内容，仅保留代码骨架与非字面量标识符。"""
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


def is_exported(name):
    """Go 规则：首字符为 Unicode 大写字母才可跨包访问。
    这里进一步限定为 ASCII 大写（该库命名风格的实际形态）。"""
    return bool(name) and 'A' <= name[0] <= 'Z'


# ---------- 包级定义（用于排除同 package 内部符号） ----------
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

# ---------- server/app/models/response 的导出名（另一处点导入） ----------
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

# ---------- 逐文件提取 ----------
syms = collections.Counter()
sym_files = collections.defaultdict(set)
sym_conf = {}                      # high / low
sym_ctx = {}                       # 调用形式样例
dot_files, alias_files = [], []
gorm_noise = {"Where", "Select", "Order", "Group", "Count", "First", "Find",
              "Updates", "Clauses", "Sum", "Pluck", "Joins", "Preload", "Model",
              "Table", "Limit", "Offset", "Delete", "Create", "Save", "Raw", "Exec"}

for p in go_files:
    raw = read(p)
    if '"EFunc/utils"' not in raw:
        continue
    rel = os.path.relpath(p, ROOT)
    code = strip_literals(raw)

    alias = "utils"
    for line in raw.splitlines():
        if '"EFunc/utils"' in line:
            m = re.match(r'\s*(' + ID + r'|\.)\s+"EFunc/utils"', line)
            alias = m.group(1) if m else "utils"
            break

    # 本文件局部声明
    local = set(re.findall(r'(?<![\w.])(' + ID + r')\s*:=', code))
    for m in re.finditer(r'(?<![\w.])((?:' + ID + r'\s*,\s*)+' + ID + r')\s*:=', code):
        for part in m.group(1).split(','):
            local.add(part.strip())
    local |= set(re.findall(r'\bvar\s+(' + ID + r')', code))
    local |= set(re.findall(r'^func\s+(?:\([^)]*\)\s*)?(' + ID + r')\s*[\(\[]', code, re.M))
    local |= set(re.findall(r'^type\s+(' + ID + r')', code, re.M))

    found = collections.Counter()
    if alias == ".":
        dot_files.append(rel)
        for m in re.finditer(r'(?<![\w.])(' + ID + r')\s*\(', code):
            name = m.group(1)
            if not is_exported(name): continue
            if name in GO_KEYWORDS or name in local: continue
            if name in pkg_defs[os.path.dirname(p)]: continue
            if name in response_exports or name in gorm_noise: continue
            found[name] += 1
            sym_conf[name] = sym_conf.get(name, "low")
            sym_ctx.setdefault(name, name + "(")
        for m in re.finditer(r'(?<![\w.])(' + ID + r')\s*\{', code):
            name = m.group(1)
            if not is_exported(name): continue
            if name in GO_KEYWORDS or name in local: continue
            if name in pkg_defs[os.path.dirname(p)]: continue
            if name in response_exports or name in gorm_noise: continue
            found[name] += 1
            sym_conf[name] = sym_conf.get(name, "low")
            sym_ctx.setdefault(name, name + "{...}")
    else:
        alias_files.append(rel)
        for m in re.finditer(r'(?<![\w.])' + re.escape(alias) + r'\.(' + ID + r')', code):
            name = m.group(1)
            found[name] += 1
            sym_conf[name] = "high"
            if re.search(re.escape(alias) + r'\.' + re.escape(name) + r'\s*\(', code):
                sym_ctx.setdefault(name, alias + "." + name + "(...)")
            else:
                sym_ctx.setdefault(name, alias + "." + name)

    for name, cnt in found.items():
        syms[name] += cnt
        sym_files[name].add(rel)

# ---------- 输出 ----------
high = [(k, v) for k, v in syms.most_common() if sym_conf.get(k) == "high"]
low = [(k, v) for k, v in syms.most_common() if sym_conf.get(k) != "high"]

print("=" * 92)
print("文件: %d 个 .go / 其中 %d 个依赖 EFunc/utils  (点导入 %d / 别名 %d)"
      % (len(go_files), len(dot_files) + len(alias_files), len(dot_files), len(alias_files)))
print("EFunc 符号: 共 %d 个   高置信(别名精确命中) %d   待核实(仅点导入) %d"
      % (len(syms), len(high), len(low)))
print("=" * 92)
print("\n【高置信符号】符号 | 引用 | 文件 | 样例调用")
for name, n in high:
    print("%-34s %5d %4d  %s" % (name, n, len(sym_files[name]), sym_ctx.get(name, "")))
if low:
    print("\n【待核实】符号 | 引用 | 文件 | 样例调用")
    for name, n in low:
        print("%-34s %5d %4d  %s" % (name, n, len(sym_files[name]), sym_ctx.get(name, "")))

with open(os.path.join(os.path.dirname(ROOT), "efunc_syms.json"), "w", encoding="utf-8") as fh:
    json.dump({
        "summary": {"go_files": len(go_files), "dependent_files": len(dot_files) + len(alias_files),
                    "dot_import_files": len(dot_files), "alias_import_files": len(alias_files),
                    "symbols_total": len(syms), "high": len(high), "low": len(low)},
        "high": [{"name": k, "calls": v, "files": len(sym_files[k]), "ctx": sym_ctx.get(k, "")} for k, v in high],
        "low": [{"name": k, "calls": v, "files": len(sym_files[k]), "ctx": sym_ctx.get(k, "")} for k, v in low],
    }, fh, ensure_ascii=False, indent=2)
print("\n明细 -> efunc_syms.json")
