package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const userLimit = 16
const privateMessageLimit = 16

var userCounter uint32
var userGroup map[string]user
var messageChannel chan message
var roomGroup map[string]room

func main() {
	var connections = make(chan net.Conn)
	messageChannel = make(chan message, 10)
	var console = make(chan string)
	userGroup = make(map[string]user)
	roomGroup = make(map[string]room)
	roomGroup["main"] = room{name: "main", users: []user{}}

	//userGroup["admin"] = user{
	//	uName:           "admin",
	//	uID:             0,
	//	connection:      nil,
	//	privateMessages: make(chan []string, privateMessageLimit),
		//currentRoom:     "",
	//}

	var chatHistory = make([]string, 0)



	fmt.Println("Starting server")

	go ManageConnections(&connections)
	go ManageConsole(&console)

	for {
		select {
		case msg := <-messageChannel:
			//tm.Clear()

			user := userGroup[msg.username]
			modMsg := fmt.Sprintf("\r[%s] %s: %s\n", user.currentRoom, user.uName, msg.m)
			chatHistory = append(chatHistory, modMsg)

			for _, u := range userGroup {
				if user.uName != u.uName{
					(u.connection).Write([]byte(fmt.Sprintf("%s[%s] %s: ", modMsg, u.currentRoom, u.uName)))
				}
			}

			fmt.Print(modMsg)
			fmt.Print("$: ")

		case com := <-console:
			if strings.TrimRight(com, "\n") == "quit" {
				return
			}
		}
	}

}

//ManageConnections listens for connections, creates user models, and notifies main thread using cons channel.
func ManageConnections(cons *chan net.Conn) {
	ln, err := net.Listen("tcp", ":8081")
	defer ln.Close()
	if err != nil {
		fmt.Println(err)
	}
	for {
		con, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("New connection")
		//con.Write([]byte("Hello"))

		go CreateUser(con)
		//*cons <- con

	}
}

//CreateUser Creates a user struct from the connection and notifies the main thread.
func CreateUser(con net.Conn) {
	rd := bufio.NewReader(con)

	con.Write([]byte("Enter a username: "))
	in, err := rd.ReadString('\n')
	in = strings.TrimRight(in, "\r\n")
	if err != nil {
		fmt.Println("Error during user creation.")
	}
	user := user{in, userCounter, con, make(chan []string, privateMessageLimit), "main"}
	userGroup[in] = user
	atomic.AddUint32(&userCounter, 1)
	go ManageUser(user)
}

//ManageUser manages the main interactive loop of each user.
func ManageUser(u user) {
	u.connection.Write([]byte(fmt.Sprintf("Server Time is: %s\n", time.Now().Format("15:04:05"))))
	r := bufio.NewReader(u.connection)
	for {
		u.connection.Write([]byte(fmt.Sprintf("[%s] You: ", u.currentRoom)))
		in, err := r.ReadString('\n')
		in = strings.TrimRight(in, "\n")

		if err != nil {
			fmt.Println("Error while receiving user input" + err.Error())
			break
		}

		switch in {
		case strings.TrimRight("/quit", "\n"):
			u.connection.Close()
		default:
			messageChannel <- message{m: in, username: u.uName}
		}
	}

}

//ManageConsole handles the administrative console for managing various chatrooms and users.
func ManageConsole(cons *chan string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("$: ")
	for {
		text, _ := reader.ReadString('\n')
		words := strings.Split(strings.TrimRight(text, "\n"), " ")

		switch words[0] {
		case "/all":
			messageChannel <- message{m: text[4:], username: "admin"}
		}


		*cons <- text
	}

}
