package main

import "net"

type user struct {
	uName           string
	uID             uint32
	connection      net.Conn
	privateMessages chan []string
	currentRoom string
}
