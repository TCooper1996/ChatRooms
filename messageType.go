package main

import "strconv"

type messageType uint64

const (
	NamePrompt messageType = iota
	ConsolePrompt
	NewConnection
	ChatMessage
	UnknownCommandError
	BadFormatError
	PrivilegeError
	MessageOverflowError
	Room404Error
	RoomAlreadyExistsError
	RoomCreated
	RoomLimitReachedError
	RoomListing
	RoomChanged
	UserListing
)

func (m messageType) toString() string {
	return strconv.Itoa(int(m))
}
