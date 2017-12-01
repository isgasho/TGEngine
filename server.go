package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// ******** 主程序 ********
// 参数说明:
//     启动服务器端: server [port]
//         eg: server 9090
func main() {
	StartServer(os.Args[1])
}

// ******** 启动服务器 ********
// 参数:
//     端口 port
func StartServer(port string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+port)
	checkError(err, "ResolveTCPAddr")
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err, "ListenTCP")
	conns := make(map[string]net.Conn)
	messages := make(chan string, 10)
	//启动服务器广播线程
	go echoHandler(&conns, messages)
	for {
		fmt.Println("Listening ...")
		conn, err := listener.Accept() //不accept到就阻塞
		checkError(err, "Accept")
		fmt.Println("Accepting ...")
		conns[conn.RemoteAddr().String()] = conn //ip和port
		go Handler(&conns, conn, messages)
	}
}

// ******** 服务器发送数据的线程 ********
// 参数:
//     连接字典 conns
//     数据通道 messages
func echoHandler(conns *map[string]net.Conn, messages chan string) {
	for {
		msg := <-messages
		fmt.Println(msg)
		for key, value := range *conns {
			fmt.Println("Send message to: ", key)
			_, err := value.Write([]byte(msg))
			if err != nil {
				fmt.Println(err.Error())
				delete(*conns, key)
			}
		}
	}
}

// ******** 服务器端接收数据线程 ********
// 参数:
//     数据连接 conn
//     通讯通道 messages
func Handler(conns *map[string]net.Conn, conn net.Conn, messages chan string) {
	username := conn.RemoteAddr().String()
	fmt.Println("Connected from: ", username)
	buf := make([]byte, 1024)
	for {
		lenght, err := conn.Read(buf)
		if checkError(err, "Connection") == false {
			conn.Close()
			break
		}
		reciveStr := string(buf[0:lenght])
		if strings.HasPrefix(reciveStr, "login_") {
			username = strings.TrimLeft(reciveStr, "login_")
			username = strings.TrimRight(username, "\r\n")
			messages <- username + ":login_successed!"
		} else {
			messages <- username + ":" + reciveStr
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
