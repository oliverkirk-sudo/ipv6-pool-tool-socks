package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	listenAddr string
	ipv6Prefix string
)

func init() {
	flag.StringVar(&listenAddr, "l", "127.0.0.1:1080", "The address to listen on")
	flag.StringVar(&ipv6Prefix, "i", "2001:470:827a::/48", "The IPv6 prefix for generating random addresses")
	rand.Seed(time.Now().UnixNano())
}
func generateRandomIPv6(prefix string) (net.IP, error) {
	_, ipv6Net, err := net.ParseCIDR(prefix)
	if err != nil {
		return nil, err
	}

	maskSize, _ := ipv6Net.Mask.Size()

	ip := make(net.IP, net.IPv6len)
	copy(ip, ipv6Net.IP)

	// 为地址的主机部分生成随机值
	for i := maskSize / 8; i < net.IPv6len; i++ {
		// 如果maskSize不能被8整除，最后一个受影响的字节需要特殊处理
		if i == maskSize/8 && maskSize%8 != 0 {
			bitMask := byte(0xFF >> (maskSize % 8))
			ip[i] = (ip[i] & ^bitMask) | (byte(rand.Intn(256)) & bitMask)
		} else {
			ip[i] = byte(rand.Intn(256))
		}
	}

	return ip, nil
}
func main() {
	flag.Parse()
	// 监听本地端口
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to set up listener: %v", err)
	}
	defer listener.Close()

	log.Println("SOCKS5 server listening on " + listenAddr)

	for {
		// 接受客户端连接
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept client: %v", err)
			continue
		}

		// 处理每个客户端的连接
		go handleClientRequest(client)
	}
}

// 检查是否支持用户名/密码认证
func supportsUsernamePasswordAuth(methods []byte) bool {
	for _, method := range methods {
		if method == 0x02 {
			return true
		}
	}
	return false
}

// 进行用户名/密码认证
func authenticate(client net.Conn) bool {
	var buf [513]byte

	// 读取用户名和密码
	_, err := client.Read(buf[:])
	if err != nil {
		log.Printf("Failed to read username and password: %v", err)
		return false
	}

	// 解析用户名和密码
	uLen := int(buf[1])
	username := string(buf[2 : 2+uLen])
	pLen := int(buf[2+uLen])
	password := string(buf[3+uLen : 3+uLen+pLen])

	// 验证用户名和密码
	validUsername := os.Getenv("SOCKS5_USERNAME")
	validPassword := os.Getenv("SOCKS5_PASSWORD")
	if username == validUsername && password == validPassword {
		client.Write([]byte{0x01, 0x00}) // 认证成功
		return true
	}

	client.Write([]byte{0x01, 0x01}) // 认证失败
	return false
}
func handleClientRequest(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	// 协议版本和认证方法
	buf := make([]byte, 258)

	_, err := client.Read(buf)
	if err != nil {
		log.Printf("Failed to get version and method: %v", err)
		return
	}

	// 检查是否支持用户名/密码认证
	if !supportsUsernamePasswordAuth(buf[1 : 1+int(buf[0])]) {
		client.Write([]byte{0x05, 0xFF}) // 无支持的认证方法
		return
	}
	client.Write([]byte{0x05, 0x02})
	if !authenticate(client) {
		return
	}

	// 只支持 SOCKS5
	if buf[0] != 0x05 {
		return
	}

	n, err := client.Read(buf)
	if err != nil {
		log.Printf("Failed to get request: %v", err)
		return
	}

	if n < 7 || buf[1] != 0x01 { // 只支持 CONNECT 请求
		return
	}

	addrType := buf[3]
	var host string
	switch addrType {
	case 0x01: // IP V4
		host = net.IPv4(buf[4], buf[5], buf[6], buf[7]).String()
	case 0x03: // 域名
		host = string(buf[5 : n-2]) // 域名是动态长度
	case 0x04: // IP V6
		host = net.IP(buf[4:20]).String()
	default:
		return
	}
	log.Println("请求地址: " + string(host))

	port := int(buf[n-2])<<8 | int(buf[n-1])

	log.Println("类型: " + strconv.Itoa(port))

	// 指定本地 IP 地址
	localIP, err := generateRandomIPv6(ipv6Prefix)
	if err != nil {
		return
	}
	localAddr := &net.TCPAddr{
		IP: localIP,
	}

	dialer := net.Dialer{
		LocalAddr: localAddr,
	}
	log.Println("使用ip: " + localIP.String())

	// 连接目标服务器
	server, err := dialer.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return
	}
	defer server.Close()

	var boundIP net.IP
	var boundPort int
	boundIP = net.ParseIP(strings.Split(listenAddr, ":")[0])
	boundPort, _ = strconv.Atoi(strings.Split(listenAddr, ":")[1])

	// 构建响应
	var response []byte
	response = append(response, 0x05) // SOCKS5版本
	response = append(response, 0x00) // 请求成功
	response = append(response, 0x00) // 保留字节

	// 添加IP地址和端口
	if ipv4 := boundIP.To4(); ipv4 != nil {
		response = append(response, 0x01) // 地址类型: IPv4
		response = append(response, ipv4...)
	} else if ipv6 := boundIP.To16(); ipv6 != nil {
		response = append(response, 0x04) // 地址类型: IPv6
		response = append(response, ipv6...)
	}

	// 端口（网络字节序）
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(boundPort))
	response = append(response, portBytes...)

	// 发送响应
	_, err = client.Write(response)
	if err != nil {
		log.Printf("Failed to send response: %v", err)
		return
	}

	// 数据转发
	go io.Copy(server, client)
	io.Copy(client, server)
}
