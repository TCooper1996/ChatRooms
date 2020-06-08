package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
)

const privateMessageLimit = 16
const serverName = "server"

var userGroup map[string]*user
var messageChannel chan message
var broadcastChannel chan string
var roomGroup map[string]*room
var signalChan chan serverSignal
var logger log.Logger
var errorsEncountered = false
var server user

func main() {
	messageChannel = make(chan message, 10)
	broadcastChannel = make(chan string)
	userGroup = make(map[string]*user)
	roomGroup = make(map[string]*room)
	signalChan = make(chan serverSignal)
	roomGroup["main"] = &room{name: "main", users: make([]string, 0)}

	server = *newUser(serverName, nil)

	fmt.Println("Starting server")
	initializeLogger()
	initializeCommands()
	go ManageConnections()
	go server.ManageUser()

	for {
		select {
		case msg := <-broadcastChannel:

			modMsg := fmt.Sprintf("[Server Broadcast]: %s", msg)
			for _, room := range roomGroup {
				cacheMessage(msg, room.name)

			}
			for _, u := range userGroup {
				u.Write(modMsg)
			}

		case msg := <-messageChannel:

			user := userGroup[msg.username]
			modMsg := fmt.Sprintf("[%s] %s: %s", user.currentRoom, user.uName, msg.m)

			cacheMessage(modMsg, msg.room)

			for _, u := range roomGroup[msg.room].Range() {
				if user != u {
					u.Write(modMsg)
				}
			}

		case sig := <-signalChan:
			switch sig {
			case Quit:
				if errorsEncountered {
					fmt.Println("Issues were encountered during runtime. See log file.")
				}
				return
			}
		}
	}

}

//ManageConnections listens for connections, creates user models, and notifies main thread using cons channel.
func ManageConnections() {
	ln, err := net.Listen("tcp", ":8081")
	defer func() { _ = ln.Close() }()
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
			err = con.Close()
			if err != nil {
				log.Println("Error closing connection after max user limit reached:" + err.Error())
			}
			continue
		}
		//fmt.Printf(adminFormatting, "New connection")

		server.Write("New Connection")

		go CreateUser(con)
	}
}

//CreateUser Creates a user struct from the connection and notifies the main thread.
func CreateUser(con net.Conn) {
	rd := bufio.NewReader(con)

	_, e := con.Write([]byte("Enter a username: "))
	if e != nil {
		log.Println("Failed to send data to user: " + e.Error())
	}

	in, err := rd.ReadString('\n')
	in = strings.TrimRight(in, "\r\n")
	if err != nil {
		log.Println("Failed to receive input from user: " + err.Error())
	}
	user := newUser(in, con)
	//userGroup[in] = user
	//roomGroup["main"].users = append(roomGroup["main"].users, user.uName)
	go user.ManageUser()
}

func cacheMessage(m string, roomName string) {
	room := roomGroup[roomName]
	err := room.chatHistory.Push(m)
	if err != nil {
		logger.Println("messageQueue.Push() was called with a string that exceeded the maximum.")
		errorsEncountered = true
	}
}

func initializeLogger() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting cwd while creating logger file: %s", err.Error())
		return
	}

	f, err := os.OpenFile(path.Join(cwd, "ChatRooms_logger"), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Printf("Error creating logger file: %s", err.Error())
		return
	}

	w := bufio.NewWriter(f)
	logger = *log.New(w, "", log.Lshortfile)

}
