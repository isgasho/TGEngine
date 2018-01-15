// loginapp/client.go
// 这里是函数的实现
// 处理登录流程,当登录成功之后联系baseapp,得到uid.
// 将uid和baseapp的地址传递给用户.

package main

import (
	"errors"
	"github.com/xtaci/kcp-go"
	"log"
	"net/rpc"
	"time"
)

func main() {
	initLog()                                        //初始化日志格式
	myLoginappRpcClient := createLoginappRpcClient() //开启Loginapp连接
	myAccount := new(Account)                        //帐号密码
	myAccount.Username = "jack"
	myAccount.Password = "1234"
	getBaseappInfo(myLoginappRpcClient, myAccount) //登录,获取baseapp信息
}

const ( //相当于枚举
	ERROR = iota //错误(无需退出)
	PANIC        //异常(需要退出)
)

func initLog() { //设定日志格式
	log.SetPrefix("【client】:")                                      //设定日志前缀
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile) //设定日志格式
}

func checkError(err error, flag int) { //检查错误
	if err != nil {
		switch flag {
		case ERROR: //错误 黄色输出 不退出
			log.Printf("%c[33m %s %s %c[0m \n", 0x1B, "【error】:", err.Error(), 0x1B)
		case PANIC: //异常 红色输出 退出
			log.Fatalf("%c[31m %s %s %c[0m \n", 0x1B, "【panic】:", err.Error(), 0x1B)
		}
	}
}

func createLoginappRpcClient() *rpc.Client { //创建与Loginapp的rpc连接
	conn, err := kcp.Dial("172.17.0.2:1234") //建立连接
	checkError(err, PANIC)
	conn.SetDeadline(time.Now().Add(time.Millisecond * 460)) //设置超时
	rpcClient := rpc.NewClient(conn)                         //建立rpc连接
	isConnected := false
	rpcClient.Call("Loginapp.ConnectionTest", 1, &isConnected) //测试连接是否成功
	if !isConnected {
		checkError(errors.New("Connected timeout or failed!"), PANIC)
	}
	log.Println("Connected to Loginapp success!")
	return rpcClient
}

type Account struct { //帐号密码结构
	Username string
	Password string
}

type BaseappInfo struct { //baseapp信息结构
	Address string
	Uid     uint64
}

func getBaseappInfo(myLoginappRpcClient *rpc.Client, myAccount *Account) *BaseappInfo { //登录,获取baseapp信息
	myBaseappInfo := new(BaseappInfo)                                             //接收baseapp信息
	err := myLoginappRpcClient.Call("Loginapp.LogIn", &myAccount, &myBaseappInfo) //登录,获取baseapp信息
	checkError(err, ERROR)
	log.Println(myBaseappInfo.Address, myBaseappInfo.Uid)
	return myBaseappInfo
}
