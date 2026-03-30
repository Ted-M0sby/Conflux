# 使用安装在 D 盘的便携版 Go（与阿里云镜像下载脚本一致）。
# 用法: . .\scripts\use-go-from-d.ps1
$goRoot = "D:\tools\go-portable\go"
$goBin = Join-Path $goRoot "bin"
if (-not (Test-Path (Join-Path $goBin "go.exe"))) {
    Write-Error "未找到 $goBin\go.exe。请先下载并解压到 D:\tools\go-portable\，参见 README「D 盘便携 Go」。"
    return
}
$env:GOROOT = $goRoot
$env:PATH = "$goBin;$env:PATH"
Write-Host "GOROOT=$env:GOROOT"
& go version
