package main

const roomLimit = 16

var roomCounter uint32

type room struct {
	name        string
	users       []string
	owner       string
	chatHistory messageQueue
}

func newRoom(name string, owner string) room {
	return room{name: name, users: make([]string, 1), owner: owner, chatHistory: messageQueue{}}
}

func (r *room) AddUser(name string) {
	u := userGroup[name]
	r.users = append(r.users, u.uName)
}
