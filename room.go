package main

import (
	"errors"
	"log"
)

const roomLimit = 16

var roomCounter uint32

type room struct {
	name        string
	users       []string
	owner       string
	chatHistory messageQueue
}

func newRoom(name string, owner string) room {
	return room{name: name, users: make([]string, 0), owner: owner, chatHistory: messageQueue{}}
}

func (r *room) AddUser(name string) {
	u := userGroup[name]
	if err := roomGroup[u.currentRoom].RemoveUser(u.uName); err != nil && u.currentRoom != r.name {
		log.Println("Failed to remove user from group it should belong to: " + err.Error())
	}

	r.users = append(r.users, u.uName)
	u.currentRoom = r.name

}

func (r *room) RemoveUser(name string) error {
	arr := r.users
	for i, u := range arr {
		if u == name {
			arr[i], arr[len(arr)-1] = arr[len(arr)-1], arr[i]
			r.users = arr[:len(arr)-1]
			return nil
		}
	}
	return errors.New("user not in room")
}

func (r room) Range() []*user {
	arr := make([]*user, len(r.users))
	for i := 0; i < len(r.users); i++ {
		arr[i] = userGroup[r.users[i]]
	}
	return arr
}
