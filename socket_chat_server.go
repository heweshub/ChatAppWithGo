package main

import (
	"fmt"
	"net"
	"time"
)

var ConnSlice map[net.Conn]*HeartBeat // 保存当前连接的客户端信息

type HeartBeat struct {
	endTime int64 // 设置过期时间
}

func handleConn(c net.Conn) {
	buffer := make([]byte, 1024)

	for {
		n, err := c.Read(buffer)
		if err != nil {
			return
		}
		if ConnSlice[c].endTime > time.Now().Unix() {
			// 更新心跳时间
			ConnSlice[c].endTime = time.Now().Add(time.Second * 5).Unix()
		} else {
			// 客户端断开连接
			fmt.Println("长时间未发送消息断开连接")
			return
		}
		// 收到客户端发送的"1，并重新恢复"1"，保持心跳
		// 心跳检测会发送"1"，如果是心跳检测就跳过本次循环
		if string(buffer[0:n]) == "1" {
			c.Write([]byte("1"))
			continue
		}
		for conn, heart := range ConnSlice {
			// 不会给当前conn发送自己的信息
			if conn == c {
				continue
			}
			// 更新连接池中的连接，若连接超时，则从字典中删除conn的key
			// 超过过期时间就会自动断开连接
			if heart.endTime < time.Now().Unix() {
				delete(ConnSlice, conn)
				conn.Close()
				fmt.Println("删除连接", conn.RemoteAddr())
				fmt.Println("现在存有连接", ConnSlice)
				continue
			}
			//
			conn.Write(buffer[0:n])
		}
	}
}

func main() {
	ConnSlice = map[net.Conn]*HeartBeat{}
	l, err := net.Listen("tcp", "127.0.0.1:8086")
	if err != nil {
		fmt.Println("服务器启动失败")
	}

	defer l.Close()

	for {
		// 客户端连接
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
		}
		fmt.Printf("Received message %s -> %s  \n", conn.RemoteAddr(), conn.LocalAddr())
		ConnSlice[conn] = &HeartBeat{
			endTime: time.Now().Add(time.Second * 5).Unix(), // 初始化过期时间
		}
		go handleConn(conn)
	}
}
