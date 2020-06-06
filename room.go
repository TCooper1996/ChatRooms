package main

type room struct {
	name        string
	users       []string
	owner       string
	chatHistory []string
	chatBytes   int
}

func newRoom(name string, owner string) room {
	return room{name: name, users: make([]string, 1), owner: owner, chatHistory: make([]string, 0), chatBytes: 0}
}
