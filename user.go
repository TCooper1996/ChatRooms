package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

const userLimit = 16

var userCounter uint32 = 0

type user struct {
	uName           string
	uID             uint32
	connection      net.Conn
	privateMessages chan []string
	currentRoom     string
	admin           bool
	writer          bufio.Writer
	reader          bufio.Reader
}

func newUser(name string, con net.Conn) *user {
	c := userCounter
	atomic.AddUint32(&c, 1)
	u := &user{uName: name, uID: c, connection: con, privateMessages: make(chan []string), currentRoom: "main"}
	userGroup[name] = u
	roomGroup["main"].AddUser(name)

	if name == serverName {
		u.writer = *bufio.NewWriter(os.Stdout)
		u.reader = *bufio.NewReader(os.Stdin)
	} else {
		u.writer = *bufio.NewWriter(u.connection)
		u.reader = *bufio.NewReader(u.connection)
	}
	return u
}

func (u *user) Write(data string, state messageType) {
	var uState string
	if autotest {
		uState = "{" + strconv.Itoa(int(state)) + "}"
	} else {
		uState = ""
	}
	u.WriteRaw(fmt.Sprintf("\r%s%s\n[%s] %s:", uState, data, u.currentRoom, u.uName))

}

func (u *user) Writef(format string, state messageType, s ...string) {
	u.Write(fmt.Sprintf(format, s), state)
}

func (u *user) WriteRaw(data string) {
	mutex.Lock()
	if _, err := u.writer.WriteString(data); err != nil {
		log.Println("Error writing to user: " + err.Error())
	}

	if err := u.writer.Flush(); err != nil {
		log.Println("Error flushing: " + err.Error())
	}
	mutex.Unlock()
}

func (u *user) WritePrompt() {
	var mType string

	if autotest {
		mType = "{" + strconv.Itoa(int(ConsolePrompt)) + "}"
	} else {
		mType = ""
	}
	u.WriteRaw(fmt.Sprintf("\r%s[%s] %s: ", mType, u.currentRoom, u.uName))

}

func (u *user) ManageUser() {

	for {
		u.WritePrompt()
		text, _ := u.reader.ReadString('\n')
		text = strings.TrimRight(text, "\r\n")

		// If the first character is a '/', it is a command. Otherwise, it is a message
		if text[0] == '/' {
			words := strings.Split(text, " ")
			if c, exists := commandMap[words[0]]; exists {
				if c.adminOnly && !u.admin {
					u.Write("This function is only available to admins.", PrivilegeError)
				} else {
					c.function(u, words)
				}
			} else {
				u.Write("Unknown command. Try /help", UnknownCommandError)
			}
		} else {
			if len([]byte(text)) > maxMessageSize {
				u.Write(fmt.Sprintf("Max message size reached (%d)", maxMessageSize), MessageOverflowError)
			} else {
				messageChannel <- newMessage(text, u.uName, u.currentRoom)
			}
		}

	}
}
