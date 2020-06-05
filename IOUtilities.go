package main

import (
	"fmt"
	"os"
)

func openHistoryFile(roomName string, write bool) *os.File {
	var flags int
	if write {
		flags = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	} else {
		flags = os.O_RDONLY
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	roomFile := fmt.Sprintf("%s/%s.txt", cwd, currentRoom)
	f, err := os.OpenFile(roomFile, flags, 0644)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	return f
}
