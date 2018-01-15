// loginapp/server.go
// 这里是函数的实现
// 处理登录流程,当登录成功之后联系baseapp,得到uid.
// 将uid和baseapp的地址传递给用户.

package main

import (
	"database/sql"            //数据库接口
	"errors"                  //自定义错误
	_ "github.com/lib/pq"     //postsql接口实现
	"github.com/xtaci/kcp-go" //kcp开源库
	"log"                     //日志库
	"net/rpc"                 //远程调用库
	"sync"                    //同步库
	"sync/atomic"             //原子计数器
)

func init() { //初始化函数,在main()前自动执行
	initLog()      //初始化日志格式
	initDatabase() //初始化数据库连接池
	uid = 0        //uid初值
}

func main() {
	createRpcServer() //创建rpc服务器
}

const ( //相当于枚举
	ERROR = iota //错误(无需退出)
	PANIC        //异常(需要退出)
)

func initLog() { //设定日志格式
	log.SetPrefix("【loginapp】:")                                    //设定日志前缀
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile) //设定日志格式
}

var myDatabase *sql.DB //数据库连接池

func initDatabase() { //初始化数据库连接池
	var err error
	myDatabase, err = sql.Open("postgres", "host='172.17.0.2' port=5432 "+
		"user=root password=1 dbname=root") //开启数据库连接
	checkError(err, PANIC)
	_, err = myDatabase.Exec("CREATE TABLE IF NOT EXISTS username_password" +
		"(username VARCHAR(80), password VARCHAR(80))") //创建帐号密码表
	checkError(err, PANIC)
	log.Println("start database success!")
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

func createRpcServer() { //创建rpc服务器
	myLoginapp := new(Loginapp)     //rpc调用的接口
	err := rpc.Register(myLoginapp) //注册rcp调用接口
	checkError(err, PANIC)
	listener, err := kcp.Listen(":1234") //监听kcp
	checkError(err, PANIC)
	defer listener.Close()
	rpc.Accept(listener) //rpc调用接口监听kcp
	conn, err := listener.Accept()
	checkError(err, ERROR)
	defer conn.Close()
}

type Loginapp int //login模块函数接口

func (t Loginapp) ConnectionTest(args int, isConnected *bool) error { //测试连接是否正常
	*isConnected = true
	return nil
}

type Account struct { //帐号数据结构
	Username string
	Password string
}

type BaseappInfo struct { //baseapp数据结构
	Address string
	Uid     uint64
}

var uid uint64             //用户uid
var usernameToUid sync.Map //用户username和uid对应关系

func (t Loginapp) LogIn(account *Account, myBaseappInfo *BaseappInfo) error { //验证帐号并返回baseapp信息
	if len(account.Username) > 80 || len(account.Password) > 80 { //帐号密码长度要小于80
		return errors.New("Username or Password should shorter than 80.")
	}
	row := myDatabase.QueryRow("SELECT username, password FROM username_password "+
		"WHERE username = $1", account.Username) //查询帐号密码
	var username string
	var password string
	row.Scan(&username, &password)
	if username == "" { //查询不到用户名,新建用户.
		stmt, err := myDatabase.Prepare("INSERT INTO username_password(username, password) " +
			"VALUES($1, $2)")
		checkError(err, ERROR)
		defer stmt.Close()
		_, err = stmt.Exec(account.Username, account.Password)
		return errors.New("There is no username, but we regist it now.")
	} else if password != account.Password { //密码错误
		return errors.New("Wrong Password!")
	} else { //登录成功
		myBaseappInfo.Address = "172.17.0.2:2234"
		_, isRegistered := usernameToUid.Load(username) //usernameToUid是线程安全的map
		if !isRegistered {
			usernameToUid.LoadOrStore(username, atomic.AddUint64(&uid, 1))
		}
		value, _ := usernameToUid.Load(username)
		myBaseappInfo.Uid = value.(uint64)
		return nil
	}
}
