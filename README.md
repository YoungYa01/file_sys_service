# file_sys_service

## 打包部署

`Windows` 上打包 `Linux` 版本
```shell
# 打开 PowerShell（管理员权限）
# 设置编译目标为 Linux x86_64
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"  # 禁用 CGO（静态编译）

# 编译生成 Linux 可执行文件
go build -o sys_server_linux main.go
```

```shell
# 查看进程
ps aux | grep sys_server
# 停止进程
kill -9 进程号
# 设置可执行权限
chmod +x sys_server
# 启动服务
nohup ./sys_server > app.log 2>&1 &
```