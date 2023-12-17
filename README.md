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
socks5://127.0.0.1:1080
```

本项目仅用于学习交流，请勿用于违法用途