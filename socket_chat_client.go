package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	server := "127.0.0.1:8086"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		Log(os.Stderr, "Fatal error:", err.Error())
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		Log("Fatal error:", err.Error())
		os.Exit(1)
	}
	Log(conn.RemoteAddr().String(), "connect success!")
	Sender(conn)
	Log("end")
}

func Sender(conn *net.TCPConn) {
	defer conn.Close()

	sc := bufio.NewReader(os.Stdin)
	// 异步协程
	go func() {
		// 返回一个channel，自动阻塞一秒钟
		t := time.NewTicker(time.Second) // 创建定时器，用来定期发送心跳包给服务器端
		defer t.Stop()

		for {
			<-t.C
			_, err := conn.Write([]byte("1"))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}()
	name := ""
	fmt.Println("请输入聊天昵称")
	fmt.Fscan(sc, &name)
	msg := ""
	buffer := make([]byte, 1024)

	_t := time.NewTimer(time.Second * 5) //创建定时器，每次服务器端发送消息就刷新时间
	defer _t.Stop()
	go func() {
		<-_t.C
		fmt.Println("服务器出现故障，断开连接")
		return
	}()

	for {
		// 心跳检测
		go func() {
			for {
				n, err := conn.Read(buffer)
				if err != nil {
					return
				}
				// 收到消息就刷新_t定时器，time.Second * 5时间到了，
				// <-_t.C就不会阻塞
				_t.Reset(time.Second * 5)

				// "1"为心跳检测，不需要打印
				if string(buffer[0:1]) != "1" {
					fmt.Println(string(buffer[0:n]))
				}
			}
		}()
		// 输入message，发送给连接服务器，服务器会把信息发送给其他客户端，并过滤发送者。
		fmt.Fscan(sc, &msg)
		i := time.Now().Format("2006-01-02 15:04:05")
		conn.Write([]byte(fmt.Sprintf("%s\n%s\t: %s\n", i, name, msg)))
	}

}

func Log(v ...interface{}) {
	fmt.Println(v...)
}
