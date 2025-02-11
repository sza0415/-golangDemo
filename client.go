package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前客户端的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return nil
	}
	client.conn = conn
	return client

}

var serverIp string
var serverPort int

func init() {
	// ./client -p 127.0.0.1 -port 8080
	// 第二个参数是默认值 "127.0.0.1" 8080
	// ./client -h 参数说明
	/*
		Usage of ./client:
		  -ip string
		        server ip address (default "127.0.0.1")
		  -p int
		        server port (default 8080)
	*/
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server ip address")
	flag.IntVar(&serverPort, "p", 8080, "server port")
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag // 客户端进入什么模式
		return true        // true意味着正常业务
	} else {
		fmt.Println("---------请输入合法业务数字-------")
		return false
	}

}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println("--------请输入聊天内容，输入exit退出-------")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("Error sending message:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Scanln(&chatMsg)
	}
}
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	fmt.Println("-------当前在线的用户：--------")
	client.SelectUsers()
	fmt.Println("-------请输入聊天对象[用户名]，exit退出----")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println("-------输入聊天信息，exit退出------")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("Error sending message:", err)
				break
			}

			chatMsg = ""
			fmt.Println("-------输入聊天信息，exit退出------")
			fmt.Scanln(&chatMsg)
		}
		fmt.Println("-------当前在线的用户：--------")
		client.SelectUsers()
		fmt.Println("-------请输入聊天对象[用户名]，exit退出----")
		fmt.Scanln(&remoteName)
	}

}

/*
	user.go
	func (this *User) SendMsg(msg string) {
		this.conn.Write([]byte(msg + "\n"))
	}、
*/
// 处理由服务器向连接写去的消息
func (client *Client) DealResponse() {
	// 将连接中的数据输出到控制台
	// 一旦client.conn有数据 就直接copy到stdout标准输出上 永久阻塞监听
	io.Copy(os.Stdout, client.conn)
	/*	相当于下面代码：

		for {
			buf := make([]byte, 1024)
			n, err := client.conn.Read(buf)
			if err != nil {

			}
			fmt.Println(buf[:n])
		}
	*/
}

func (client *Client) Run() {
	for client.flag != 0 { // 用户的客户端模式未退出
		for client.menu() != true { // 合法业务数字才结束循环
		}
		switch client.flag {
		case 1:
			fmt.Println("-----1.公聊模式-----")
			client.PublicChat()
		case 2:
			fmt.Println("-----2.私聊模式-----")
			client.PrivateChat()
		case 3:
			fmt.Println("-----3.更新用户名-----")
			client.UpdateName()
		case 0:
			fmt.Println("退出客户端")
		}
	}
}
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>请输入用户名：")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("Error connecting to server")
		return
	}
	fmt.Println("Connecting to server")
	// 单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()
	client.Run()
}
