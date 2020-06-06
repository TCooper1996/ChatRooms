package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const userLimit = 16
const privateMessageLimit = 16
const messageChannelLimit = 32
const chatHistoryThreshold = 32
const admin = "admin"
const adminFormatting = "\r%s\n$: "
const userFormatting = "\r%s\n[%s] %s: "

//var userCounter uint32

var userGroup map[string]user
var messageChannel chan message
var broadcastChannel chan string
var roomGroup map[string]*room
var signalChan chan serverSignal
var currentRoom string

func main() {
	var connections = make(chan net.Conn)
	messageChannel = make(chan message, 10)
	broadcastChannel = make(chan string)
	//var console = make(chan string)
	userGroup = make(map[string]user)
	roomGroup = make(map[string]*room)
	signalChan = make(chan serverSignal)
	roomGroup["main"] = &room{name: "main", users: make([]string, 0), chatHistory: make([]string, 0)}
	currentRoom = "main"

	userGroup["admin"] = user{
		uName:           "admin",
		uID:             0,
		connection:      nil,
		privateMessages: make(chan []string, privateMessageLimit),
		currentRoom:     "main",
	}

	fmt.Println("Starting server")

	initializeCommands()
	go ManageConnections(&connections)
	go userGroup["admin"].ManageUser()

	for {
		select {
		case msg := <-broadcastChannel:

			for _, room := range roomGroup {
				modMsg := fmt.Sprintf("[Server Broadcast]: %s", msg)
				adminMessage := formatMessageAdmin(modMsg)
				logMessage := formatMessageLog(modMsg)

				logToFile(room.name, logMessage)

				for _, u := range userGroup {
					if u.uName == admin {
						fmt.Print(adminMessage)
					} else {
						userMessage := formatMessageUser(modMsg, u.uName)
						u.connection.Write([]byte(userMessage))
					}
				}
			}

		case msg := <-messageChannel:

			user := userGroup[msg.username]
			modMsg := fmt.Sprintf("[%s] %s: %s", user.currentRoom, user.uName, msg.m)
			adminMessage := formatMessageAdmin(modMsg) //, userMessage, logMessage := formatMessage(modMsg)
			logMessage := formatMessageLog(modMsg)

			logToFile(msg.room, logMessage)

			for _, u := range roomGroup[msg.room].users {
				if user.uName != u {
					userMessage := formatMessageUser(modMsg, u)
					userGroup[u].connection.Write([]byte(userMessage))
				}
			}

			if currentRoom == msg.room && user.uName != admin {
				fmt.Print(adminMessage)

			}

		case sig := <-signalChan:
			switch sig {
			case Quit:
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
		if userCounter >= userLimit {
			fmt.Printf("Rejecting incoming connection, user limit reached %d", userLimit)
			con.Close()
			continue
		}
		fmt.Printf(adminFormatting, "New connection")

		go CreateUser(con)
	}
}

//CreateUser Creates a user struct from the connection and notifies the main thread.
func CreateUser(con net.Conn) {
	rd := bufio.NewReader(con)

	con.Write([]byte("Enter a username: "))
	in, err := rd.ReadString('\n')
	in = strings.TrimRight(in, "\r\n")
	if err != nil {
		log.Fatalln("Error during user creation: " + err.Error())
	}
	user := newUser(in, con)
	userGroup[in] = user
	roomGroup["main"].users = append(roomGroup["main"].users, user.uName)
	go user.ManageUser()
}

func formatMessageAdmin(m string) string {
	return fmt.Sprintf(adminFormatting, m) //, fmt.Sprintf(userFormatting, m) + "[%s] %s: ", fmt.Sprintf(m + string('\n'))
}

func formatMessageLog(m string) string {
	return m + string('\n')
}

func formatMessageUser(m string, user string) string {
	userStruct := userGroup[user]
	return fmt.Sprintf(userFormatting, m, userStruct.currentRoom, userStruct.uName)
}

//logToFile adds the message to the cached history, and if the size of the history is beyond a constant threshold,
//opens up the file containing the history of the current channel, writes the cached data to it, and empties the cache.
func logToFile(name string, logMessage string) {
	room := roomGroup[name]
	room.chatHistory = append(room.chatHistory, logMessage)
	room.chatBytes += len([]byte(logMessage))
	if room.chatBytes >= chatHistoryThreshold {
		fmt.Printf(adminFormatting, "Writing chat history to disk.")
		f := openHistoryFile(room.name, true)
		for _, line := range room.chatHistory {
			f.WriteString(line)
		}
		f.Close()
		room.chatHistory = make([]string, 0)
		room.chatBytes = 0
	}

}
