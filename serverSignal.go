package main

type serverSignal string

const (
	//Quit signals that the app is attempting to quit. Goroutines that read this must wrap up and end
	Quit serverSignal = "Quit"
)
