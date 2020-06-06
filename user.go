package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
)

type user struct {
	uName           string
	uID             uint32
	connection      net.Conn
	privateMessages chan []string
	currentRoom     string
	admin           bool
}

var userCounter uint32 = 0

func newUser(name string, con net.Conn) user {
	c := userCounter
	atomic.AddUint32(&c, 1)
	return user{uName: name, uID: c, connection: con, privateMessages: make(chan []string), currentRoom: "main"}
}

func (u user) Write(data string) {
	if u.uName == admin {
		fmt.Printf("\r%s\n[%s] %s: ", data, u.currentRoom, u.uName)
	} else {

		byteData := []byte(fmt.Sprintf("\r%s\n[%s] %s: ", data, u.currentRoom, u.uName))

		bytesWritten, err := u.connection.Write(byteData)
		if err != nil {
			fmt.Println(err.Error())
		} else if bytesWritten != len(byteData) {
			fmt.Println("Error, not all bytes written to client.")
		}
	}
}

func (u user) ManageUser() {
	reader := bufio.NewReader(os.Stdin)
	if u.uName == admin {
		fmt.Printf("\r[%s] %s: ", u.currentRoom, u.uName)
	} else {
		u.connection.Write([]byte(fmt.Sprintf("\r[%s] %s: ", u.currentRoom, u.uName)))
	}

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimRight(text, "\n")

		// If the first character is a '/', it is a command. Otherwise, it is a message
		if text[0] == '/' {
			words := strings.Split(text, " ")
			if c, exists := commandMap[words[1]]; exists {
				if c.adminOnly && !u.admin {
					u.Write("This function is only available to admins.")
				} else {
					c.function(&u, words)
				}
			} else {
				u.Write("Unknown command. Try /help")
			}
		} else {
			messageChannel <- newMessage(text, u.uName, u.currentRoom)
		}

	}
}
