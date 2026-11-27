@echo off
chcp 65001 >nul
rem ============================================================
rem  飞鸟快验 Windows 版 一键构建脚本
rem
rem  做的事：
rem    1. 用 goversioninfo 把图标+版本信息编译成 resource_windows_amd64.syso
rem    2. go build 出 exe（必须用 "go build ."，不能写 "go build main.go"）
rem
rem  注意：go build main.go 只会编译指定的那一个文件，
rem        同目录下的 .syso 不会被链接，图标就丢了。
rem ============================================================
setlocal

set ROOT=%~dp0
set KY=%ROOT%ky
set DIST=%ROOT%dist

if "%GOROOT%"=="" (
    echo [错误] 未检测到 GOROOT，请先安装 Go 1.25+ 或手动设置 GOROOT
    exit /b 1
)

set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

echo [1/3] 生成图标与版本资源...
pushd "%KY%"
if not exist "app\resource\app.ico" (
    echo [错误] 找不到图标文件 app\resource\app.ico
    popd & exit /b 1
)
goversioninfo -o resource_windows_amd64.syso versioninfo.json
if errorlevel 1 (
    echo [错误] goversioninfo 生成失败，请先执行：
    echo        go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
    popd & exit /b 1
)

echo [2/3] 编译...
if not exist "%DIST%" mkdir "%DIST%"
go build -o "%DIST%\feiniao.exe" .
if errorlevel 1 (
    echo [错误] 编译失败
    popd & exit /b 1
)
popd

echo [3/3] 完成
for %%F in ("%DIST%\feiniao.exe") do echo      %%~zF 字节  %DIST%\feiniao.exe
endlocal
