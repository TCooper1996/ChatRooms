package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
)

type command struct {
	name        string
	description string
	format      string
	adminOnly   bool
	function    func(*user, []string)
}

var commandMap map[string]command

func initializeCommands() {
	commandMap = map[string]command{
		"/all": {
			name:        "/all",
			description: "Sends a message to every user regardless of room.",
			format:      "/all [message]",
			adminOnly:   true,
			function:    all,
		},

		"/create": {
			name:        "/create",
			description: "Creates a new room.",
			format:      "/create [room name]",
			adminOnly:   false,
			function:    create,
		},

		"/switch": {
			name:        "/switch",
			description: "Switches to another room.",
			format:      "/switch [room name]",
			adminOnly:   false,
			function:    switchRoom,
		},

		"/list": {
			name:        "/list",
			description: "Lists the name and number of users of each room.",
			format:      "/list",
			adminOnly:   false,
			function:    list,
		},

		"/help": {
			name:        "/help",
			description: "Displays information on all commands.",
			format:      "/help",
			adminOnly:   false,
			function:    help,
		},
	}
}
func all(_ *user, words []string) {
	broadcastChannel <- strings.Join(words[1:], " ")
}

func create(u *user, words []string) {
	//Fail if incorrect invocation format
	if len(words) != 2 {
		u.Write(fmt.Sprintf("Invalid format. Format is: \n%s.", commandMap[words[0]].format))
		//Fail if room limit reached
	} else if roomCounter > roomLimit {
		u.Write(fmt.Sprintf("Maximum number of rooms (%d) reached.", roomCounter))
		//Fail if room exists
	} else if _, exists := roomGroup[words[1]]; exists {
		u.Write("Room already exists.")
		//Success
	} else {
		r := newRoom(words[1], u.uName)
		roomGroup[words[1]] = &r
		atomic.AddUint32(&roomCounter, 1)
	}
}

func switchRoom(u *user, words []string) {
	if len(words) != 2 {
		u.Write(fmt.Sprintf("Invalid format. Format is: \n%s.", commandMap[words[0]].format))
		return
	}

	if _, exists := roomGroup[words[1]]; exists {
		u.currentRoom = words[1]
		u.Write(fmt.Sprintf("Entering room %s", u.currentRoom))

		f := openHistoryFile(u.currentRoom, false)
		s := bufio.NewScanner(f)
		//todo: Consider sending more than just one line per packet
		for s.Scan() {
			u.Write(s.Text())
		}
		roomGroup[u.currentRoom].users = append(roomGroup[u.currentRoom].users, u.uName)
	} else {
		u.Write("Room does not exist.")
	}
}

func list(u *user, _ []string) {
	//Todo: Send more than one line at a time
	u.Write("Name\tUsers")
	for _, r := range roomGroup {
		u.Write(fmt.Sprintf("%s\t%d", r.name, len(r.users)))
	}
}

func help(u *user, _ []string) {
	var str strings.Builder
	for _, c := range commandMap {
		if !(c.adminOnly && !u.admin) {
			str.WriteString(fmt.Sprintf("Name: %s\n\tDescription: %s\n\tFormat: %s", c.name, c.description, c.format))
		}
	}

	bytes := []byte(str.String())
	_, err := u.connection.Write(bytes)
	if err != nil {
		log.Println("Error sending data to user: ", err.Error())
	}

}
