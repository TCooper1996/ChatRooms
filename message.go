package main

type message struct {
	m        string
	username string
	room     string
}

func newMessage(m string, user string, room string) message {
	return message{m, user, room}
}
