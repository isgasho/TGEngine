package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

// ******** 主程序 ********
// 参数说明:
//     启动客  户端: client [Server Ip Addr]:[Server Port]
//         eg: client 192.168.0.74:9090
func main() {
	StartClient(os.Args[1])
}

// ******** 客户端启动函数 ********
// 参数: 远程ip地址和端口 tcpaddr
func StartClient(tcpaddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", tcpaddr)
	checkError(err, "ResolveTCPAddr")
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err, "DialTCP")
	//启动客户端发送线程
	go chatSend(conn)
	//开始客户端轮训
	buf := make([]byte, 1024)
	for {
		lenght, err := conn.Read(buf)
		timeSnap := time.Now().UnixNano() //时间戳
		fmt.Println(timeSnap / 100000)
		if checkError(err, "Connection") == false {
			conn.Close()
			fmt.Println("Server is dead ...ByeBye")
			os.Exit(0)
		}
		fmt.Println(string(buf[0:lenght]))
	}
}

// ******** 客户端发送线程 ********
// 参数:
//     发送连接 conn
func chatSend(conn net.Conn) {
	var input string
	for {
		fmt.Scanf("%s", &input)
		if input == "/quit" {
			fmt.Println("ByeBye..")
			conn.Close()
			os.Exit(0)
		}
		_, err := conn.Write([]byte(input + "\r\n"))
		timeSnap := time.Now().UnixNano() //时间戳
		fmt.Println(timeSnap / 100000)
		input = ""
		if err != nil {
			fmt.Println(err.Error())
			conn.Close()
			break
		}
	}
}

// ******** 错误检查 ********
func checkError(err error, info string) (res bool) {
	if err != nil {
		fmt.Println(info + "  " + err.Error())
		return false
	}
	return true
}
