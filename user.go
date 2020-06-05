package main

import (
	"net"
	"sync/atomic"
)

type user struct {
	uName           string
	uID             uint32
	connection      net.Conn
	privateMessages chan []string
	currentRoom     string
}

var userCounter uint32 = 0

func newUser(name string, con net.Conn) user {
	c := userCounter
	atomic.AddUint32(&c, 1)
	return user{uName: name, uID: c, connection: con, privateMessages: make(chan []string), currentRoom: "main"}
}
