# Ipv6PoolToolSocks

看完[自选ipv6出口](https://zu1k.com/posts/tutorials/http-proxy-ipv6-pool/)代理后写了个工具用于支持socks5请求

我使用的是[tunnelbroker](https://www.tunnelbroker.net/)提供的ipv6地址池，/48的地址池几辈子用不完

## 使用

```shell
go build
chmod +x proxytool
./proxytool -i "127.0.0.1:1080" -l "xxxx:xxxx:xxxx:xxxx::/xx"
```
接受两个参数，-i为绑定ip，-l为ipv6段

环境变量中设置
```shell
export SOCKS5_USERNAME=username
export SOCKS5_PASSWORD=password
```

设置代理
```shell
socks5://username:password@127.0.0.1:1080
```

## 注册为系统服务
在/etc/systemd/system目录中创建proxytool.service，并填入以下内容

```shell
[Unit]
Description=IPV6 Proxy Pool Tool
After=network.target

[Service]
ExecStart=/usr/bin/proxytool -i xxxx:xxxx:xxxx::/48 -l 127.0.0.1:1080
Restart=on-failure
Environment="SOCKS5_USERNAME=username" "SOCKS5_PASSWORD=password"
[Install]
WantedBy=multi-user.target

```
`/usr/bin/proxytool`替换为你的proxytool地址
保存后执行
```shell
sudo systemctl daemon-reload
sudo systemctl start proxytool
sudo systemctl enable proxytool #开机自启
sudo systemctl status proxytool #查看状态
```

本项目仅用于学习交流，请勿用于违法用途