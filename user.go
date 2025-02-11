package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn // 与客户端通信的连接
	server *Server  // 用户所属于的serve地址
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

// 用户的上线功能
func (this *User) Online() {

	// 当前用户上线了 将用户加入到onlineMap表中
	this.server.mapLock.Lock()
	this.server.onlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线通知
	this.server.Broadcast(this, "已上线")

}

// 用户的下线功能
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.onlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户上线通知
	this.server.Broadcast(this, "已下线")
}

// 给当前的用户客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\n"))
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		this.server.mapLock.Lock()
		for _, user := range this.server.onlineMap {
			//user.C <- user.Name
			this.SendMsg("用户：" + user.Name + "在线")
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 rename|张三
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.onlineMap[newName]
		if ok {
			this.SendMsg("当前用户名已有人占用")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.onlineMap, this.Name)
			this.server.onlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.SendMsg("用户名" + this.Name + "已经更新为" + newName)
			this.Name = newName
		}
	} else if len(msg) >= 3 && msg[:3] == "to|" {
		// 消息格式 to|张三｜消息内容

		// 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请按照to|张三|消息内容")
			return
		}
		remoteUser, ok := this.server.onlineMap[remoteName]
		if !ok {
			this.SendMsg("用户名不存在，请重新输入")
			return
		}
		if parts := strings.Split(msg, "|"); len(parts) != 3 {
			this.SendMsg("消息格式不正确，请按照to|张三|消息内容")
			return
		} else {
			content := strings.Split(msg, "|")[2]
			if content == "" {
				this.SendMsg("你发送的消息是空消息")
				return
			} else {
				remoteUser.SendMsg("[用户" + this.Name + "]" + "--->You:" + content)
			}

		}

	} else {
		this.server.Broadcast(this, msg)
	}

}

// 监听当前的user的channel 一旦有消息 就直接发给对应的客户端
func (this *User) ListenMessage() {
	for {
		select {
		case msg := <-this.C: // 如果channel中有数据
			this.conn.Write([]byte(msg + "\n"))
		}
	}
}
