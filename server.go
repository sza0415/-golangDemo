package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	ip   string
	port int

	// 在线用户列表
	onlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	return &Server{ip: ip, port: port, onlineMap: make(map[string]*User), Message: make(chan string)}
}

// 写一个广播方法 由新上线的用户对所有当前在线的用户进行广播
func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := "（广播）用户 " + user.Name + "]:" + msg
	this.Message <- sendMsg

}

// 监听message的channel中的消息 一旦有消息就发送给全部的在线用户
func (this *Server) ListenMessage() {
	for {
		select {
		case msg := <-this.Message:
			this.mapLock.Lock()
			for _, u := range this.onlineMap {
				u.C <- msg
			}
			this.mapLock.Unlock()
		}
	}

}

// 当前连接的业务
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("[与服务器连接建立成功,连接地址为]:" + conn.RemoteAddr().String())
	// 根据连接创建用户 并绑定服务器的地址
	/*
		type User struct {
			Name   string
			Addr   string
			C      chan string
			conn   net.Conn // 与客户端通信的连接
			server *Server  // 用户所属于的serve地址
		}
	*/
	user := NewUser(conn, this)

	user.Online()
	islive := make(chan bool)
	// 开启一个协程 用来接受客户端发送的消息
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Println("conn read err:", err)
				return
			}
			if n == 0 {
				user.Offline()
				return
			}
			// 提取用户的消息 去掉最后的"\n"
			msg := string(buf[:n-1])
			user.DoMessage(msg)
			islive <- true
		}
	}()

	for {
		select {
		case <-islive:

		case <-time.After(time.Second * 100):
			user.SendMsg("你被踢了")
			//user.Offline()

			close(user.C)
			conn.Close()
			return
		}
	}

}

// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", this.ip+":"+strconv.Itoa(this.port))

	if err != nil {
		fmt.Println("net listen err")
		return
	}
	// 启动监听 广播的message
	go this.ListenMessage()

	for {
		// accpet
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept err")
			continue
		}
		// 每次新连接就是一个业务
		go this.Handler(conn)
	}

	// close listen socket
	defer listener.Close()
}
