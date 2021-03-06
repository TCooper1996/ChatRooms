package main

import (
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

		"/listChannels": {
			name:        "/listChannels",
			description: "Lists the name and number of users of each room.",
			format:      "/listChannels",
			adminOnly:   false,
			function:    listChannels,
		},

		"/listUsers": {
			name:        "/listUsers",
			description: "Lists the name and current room of each user",
			format:      "/listUsers",
			adminOnly:   false,
			function:    listUsers,
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
		u.Writef("Invalid format. Format is: %s.", BadFormatError, commandMap[words[0]].format)
		//Fail if room limit reached
	} else if roomCounter > roomLimit {
		u.Writef("Maximum number of rooms (%s) reached.", RoomLimitReachedError, string(roomCounter))
		//Fail if room exists
	} else if _, exists := roomGroup[words[1]]; exists {
		u.Write("Room already exists.", RoomAlreadyExistsError)
		//Success
	} else {
		r := newRoom(words[1], u.uName)
		roomGroup[words[1]] = &r
		atomic.AddUint32(&roomCounter, 1)
		u.Writef("Room %s created.", RoomCreated, words[1])
	}
}

func switchRoom(u *user, words []string) {
	if len(words) != 2 {
		u.Write(fmt.Sprintf("Invalid format. Format is: \n%s.", commandMap[words[0]].format), BadFormatError)
		return
	}

	if r, exists := roomGroup[words[1]]; exists {
		r.AddUser(u.uName)

		u.Writef("Entering room %s\n", RoomChanged, u.currentRoom)

		//todo: Consider sending more than just one line per packet
		for _, m := range r.chatHistory.Range() {
			u.WriteRaw(m)
		}
		//u.WritePrompt()

	} else {
		u.Write("Room does not exist.", Room404Error)
	}
}

func listChannels(u *user, _ []string) {
	//Todo: Send more than one line at a time
	u.Write("Name\tUsers", RoomListing)
	for _, r := range roomGroup {
		u.Writef("%s\t%d", RoomListing, r.name, string(len(r.users)))
	}
}

func listUsers(u *user, _ []string) {
	u.Write("Name\tRoom", UserListing)
	for _, usr := range userGroup {
		u.Writef("%s\t%s", UserListing, usr.uName, usr.currentRoom)
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
