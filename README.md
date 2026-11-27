# 飞鸟快验 · Windows 维护版

自建的 Windows 编译/运行分支，基于官方源码补齐了缺失依赖，使其能在 Windows 上直接 `go build` 出可执行文件并运行。

| 项 | 说明 |
|---|---|
| 上游后端 | [x1056795985/KuaiYanGo2](https://github.com/x1056795985/KuaiYanGo2) v1.0.475 |
| 上游前端 | [x1056795985/KuaiYanWeb](https://github.com/x1056795985/KuaiYanWeb)（仅管理员后台） |
| 二进制参照 | 官方 1.0.548 Linux 版（`.bin`，UPX 压缩的 ELF） |
| 产物形态 | 纯源码仓库，`go build` 自行产出 Windows 单文件可执行程序（约 80.9 MB） |

> 这是**服务端后台程序**（网络验证 / 软件授权系统），默认监听 `18888`。不是桌面软件，需要 MySQL 5.7+。

---

## 一、相对上游做了什么

上游源码拿下来是**编译不过的**，原因有三处，本仓库逐一补齐：

### 1. EFunc 私有模块 → 自实现

上游 `go.mod` 写着：

```
require EFunc v0.0.0
replace EFunc => E:\yun\project\EFunc
```

`EFunc` 是作者本机私有工具库，GitHub / Gitee 都没公开，17 个 fork 也无一上传，仓库内无 vendor。而 102 个源文件全部 `import "EFunc/utils"`（其中 64 个是点导入的裸调用）。

**做法**：扫描全部调用点反推签名，按 `EFunc/utils/*.go` 分文件重写，现已 70+ 函数、编译零错误。

关键函数的行为不是猜的 —— 官方 `.bin` 没有 strip，`go tool nm` 能列出全部中文函数名，`go tool objdump` 反汇编会带出原版源文件名和行号，据此逐条还原（详见 [`编译说明.md`](编译说明.md)）。

### 2. 三套前端资源 → 补齐

上游 `app/web/{VueAdmin,VueAgent,webUser}` 里**只有 `embed.go`，一个静态资源都没有**，`//go:embed` 直接报错。

| 前端 | 来源 | 说明 |
|---|---|---|
| `VueAdmin` | KuaiYanWeb 源码 `vite build` | 作者把 outDir 写死成自己本机路径，已改相对路径 |
| `VueAgent` | 从官方 `.bin` 提取 | 作者未开源代理端 |
| `webUser` | 从官方 `.bin` 提取 | 同上，uni-app 产物 |

提取方法：`upx -d` 解包得 78 MB 未压缩 ELF，Go 的 `//go:embed` 会把文件表编译成 `{name string; data string; hash[16]byte}` 连续结构体，文件名字符串的虚址即 `name.ptr`，据此反查 `(ptr,len)` 定位 `data.ptr/data.len` 取原始字节。379/381 成功，8.6 MB，已逐个校验与 HTTP 返回字节数一致。

脚本：[`extract_embed.py`](extract_embed.py)

### 3. go.mod replace → 指向本地

```go
replace EFunc => ../EFunc
```

---

## 二、目录结构

```
.
├── ky/                     后端源码（已改 go.mod 指向 ../EFunc）
│   └── app/web/            三套前端资源已就位
├── EFunc/                  自实现的 EFunc/utils 包
├── web/                    管理员前端源码（KuaiYanWeb）
├── dist/
│   └── config.json         配置模板（数据库、端口、入口名）
├── 编译说明.md              详细文档：还原过程、差异对照、已知限制
├── extract_embed.py        前端资源提取脚本
├── scan_efunc*.py          EFunc 符号扫描脚本（改依赖时重跑）
├── EFunc待实现清单.md       符号清单与签名推断依据
└── efunc_syms.json         扫描结果原始数据
```

不进仓库的东西：Go 工具链与模块缓存、`web/node_modules`、`dist` 下的可执行文件。这些均可重新获取，本机已清理，编译前装好 Go 1.25+ 即可。

---

## 三、编译

需要 Go 1.25+（本机验证版本 1.25.14）、Node 18+（仅重建管理员前端时需要）。

```powershell
$env:CGO_ENABLED = "0"
$env:GOPROXY = "https://goproxy.cn,direct"
cd ky
go build -o ..\dist\feiniao.exe main.go
```

纯 Go、无 cgo、无平台专用文件（`_linux.go` 之类一个都没有），Windows / Linux 都能直接编。

### 重建管理员前端（可选）

```powershell
cd web
npm install
npm run build
Copy-Item -Recurse -Force dist\* ..\ky\app\web\VueAdmin\
```

代理端和用户中心**无法重建**，源码不在作者手里公开，只能用已提取的产物。

---

## 四、运行

1. 改 `dist/config.json` 的 `mysql` 段指向你的库，库可以留空，表会自动建
2. 执行 `feiniao.exe`

```
后台端口: :18888
管理员后台: http://127.0.0.1:18888/Admin/
代理平台:   http://127.0.0.1:18888/Agent/
用户中心:   http://127.0.0.1:18888/user/10001/
```

首次启动若没有 `config.json` 会在控制台提示输入 MySQL 信息。

> 用户中心地址是 `/user/{AppId}/`，**AppId 不是固定的**，同一套程序可服务任意多个应用。前端 JS 会用 `location.pathname` 里 `user/` 与下一个 `/` 之间的数字作为 AppId 发起请求，所以换个 AppId 换个 URL 就能进对应应用的界面。限制只有三条：AppId ≥ 10000、库里存在该应用、未停止运营。
>
> `index.html` 与 JS 里的静态资源前缀写死为 `/user/10001/assets/`，但静态路由同样接受任意 AppId，因此换 AppId 后资源照常加载，只是请求路径仍显示 10001，属界面外的观感问题，不影响功能。

---

## 五、已知限制

### Windows 上两个功能必然失效

源码自己写了 `if runtime.GOOS == "windows"` 直接拒绝：

- **热重启**（`utils.Reload`）→ `return errors.New("系统不支持")`
- **在线更新**（Updater）→ 硬编码宝塔路径 `/www/server/panel/data/default.db` + `os.Chmod(0755)` + `kill -1`

后台"在线更新"按钮点了会失败，属预期。

### 版本号与前端存在跨代

- 源码写死 v1.0.475，产物显示也是这个
- 前端资源来自 1.0.548

若某接口报错，优先怀疑前后端版本错位。

### 语义细节

金额舍入、脱敏等易错函数已从二进制反汇编确认并逐条对齐，剩余推断项与两处主动偏离（除零、空串脱敏，均为避免 panic）列在 [`编译说明.md`](编译说明.md) 第三节。

---

## 六、同步上游更新

本仓库的 `ky/` 和 `web/` 已移除上游 `.git`（归档在本机 `.git-archived/`），不跟踪上游历史。需要跟进官方版本时：

```bash
git clone --depth 1 https://github.com/x1056795985/KuaiYanGo2.git /tmp/upstream
diff -r /tmp/upstream/app ky/app | grep -v "app/web"    # 前端目录会大量噪音，跳过
```

重点比对 `app/logic/` 与 `app/controller/`，把业务改动手动搬过来；`EFunc/utils/` 若上游新增函数调用，重跑 `scan_efunc3.py` 生成新的待实现清单。
