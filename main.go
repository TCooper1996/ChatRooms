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

var userCounter uint32
var userGroup []user
var msgs chan string

func main() {

	var connections = make(chan net.Conn)
	msgs = make(chan string, 10)
	//var aconns = make(map[net.Conn]int)
	var console = make(chan string)
	userGroup = make([]user, userLimit)
	//var i int

	fmt.Println("Starting server")

	go ManageConnections(&connections)
	go ManageConsole(&console)

	for {
		select {
		case msg := <-msgs:
			for _, u := range userGroup {
				fmt.Println()
				(u.connection).Write([]byte(msg))
				fmt.Println("Message sent")
			}

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

	con.Write([]byte("Enter a username.\n"))
	in, err := rd.ReadString('\n')
	if err != nil {
		fmt.Println("Error during user creation.")
	}
	user := user{in, userCounter, con, make(chan string, 5)}
	userGroup = append(userGroup, user)
	atomic.AddUint32(&userCounter, 1)
	go ManageUser(user)
}

//ManageUser manages the main interactive loop of each user.
func ManageUser(u user) {
	u.connection.Write([]byte(fmt.Sprintf("Server Time is: %s\n", time.Now().Format("15:04:05"))))
	r := bufio.NewReader(u.connection)
	for {
		in, err := r.ReadString('\n')
		if err != nil {
			fmt.Println("Error while receiving user input" + err.Error())
			break
		}

		switch in {
		case strings.TrimRight("/quit", "\n"):
			u.connection.Close()
		default:
			msgs <- in
		}
	}

}

//ManageConsole handles the administrative console for managing various chatrooms and users.
func ManageConsole(cons *chan string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Console active")
	for {
		text, _ := reader.ReadString('\n')
		*cons <- text
	}

}
