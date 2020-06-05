package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const userLimit = 16
const privateMessageLimit = 16
const chatHistoryThreshold = 32
const admin = "admin"
const adminFormatting = "\r%s\n$: "
const userFormatting = "\r%s\n"

var userCounter uint32
var userGroup map[string]user
var messageChannel chan message
var roomGroup map[string]*room
var signalChan chan serverSignal
var currentRoom string

func main() {
	var connections = make(chan net.Conn)
	messageChannel = make(chan message, 10)
	var console = make(chan string)
	userGroup = make(map[string]user)
	roomGroup = make(map[string]*room)
	signalChan = make(chan serverSignal)
	roomGroup["main"] = &room{name: "main", users: make([]string, 0), chatHistory: make([]string, 0)}
	currentRoom = "main"

	//userGroup["admin"] = user{
	//	uName:           "admin",
	//	uID:             0,
	//	connection:      nil,
	//	privateMessages: make(chan []string, privateMessageLimit),
	//	currentRoom:     "main",
	//}

	fmt.Println("Starting server")

	go ManageConnections(&connections)
	go ManageConsole(&console)

	for {
		select {
		case msg := <-messageChannel:

			user := userGroup[msg.username]
			modMsg := fmt.Sprintf("[%s] %s: %s", user.currentRoom, user.uName, msg.m)
			adminMessage := fmt.Sprintf(adminFormatting, modMsg)
			userMessage := fmt.Sprintf(userFormatting, modMsg)
			logMessage := fmt.Sprintf(modMsg + string('\n'))
			room := roomGroup[msg.room]
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

			for _, u := range userGroup {
				if user.uName != u.uName {
					(u.connection).Write([]byte(userMessage))
				}
			}

			if currentRoom == msg.room {
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
		fmt.Printf(adminFormatting, "New connection")
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
		log.Fatalln("Error during user creation: " + err.Error())
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
			log.Fatalln("Error while receiving user input: " + err.Error())
		}
		fmt.Print(in)
		switch in {

		case strings.TrimRight("/quit", "\n"):
			u.connection.Close()
		default:
			messageChannel <- message{m: in, username: u.uName, room: u.currentRoom}
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
			messageChannel <- message{m: text[4:], username: admin}
		case "/quit":
			for _, u := range userGroup {
				if u.connection != nil {
					u.connection.Close()
				}
			}
			signalChan <- Quit

		case "/create":
			var roomName string
			if len(words) == 2 {
				roomName = words[1]
			} else {
				var err error
				roomName, err = reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error: " + err.Error())
				}
			}

			room := room{name: roomName, users: make([]string, 0), owner: admin}
			roomGroup[roomName] = &room
			fmt.Printf("\rRoom %s created\n$: ", roomName)

		case "/switch":
			var roomName string
			if len(words) == 2 {
				roomName = words[1]
			} else {
				var err error
				roomName, err = reader.ReadString('\n')
				if err != nil {
					log.Println("Error: " + err.Error())
				}
			}

			if r, exists := roomGroup[roomName]; exists {
				currentRoom = r.name
				fmt.Printf("\rEntering room %s\n$: ", roomName)
				f := openHistoryFile(currentRoom, false)
				s := bufio.NewScanner(f)
				for s.Scan() {
					fmt.Println(s.Text())
				}

			} else {
				fmt.Printf(adminFormatting, "Room does not exist. Create with /create [room_name]")
			}
		case "/memcount":
			fmt.Sprintf(adminFormatting, string(roomGroup[currentRoom].chatBytes))
		}

	}

}
