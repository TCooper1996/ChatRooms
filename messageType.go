package main

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
	UserListing
)
