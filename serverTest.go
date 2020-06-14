package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var randomMessages = []string{
	"hello",
	"I'm new here.",
	"I want to have a good time.",
	"Tell me about yourself.",
}

var userNames = []string{
	"Tom",
	"Michael",
	"Sharice",
	"Diamond",
	"Tayler",
	"Pedro",
	"Jerome",
	"Tamika",
}

var regex, _ = regexp.Compile("{[0-9+]}")

func userTestLoop(name string) {
	time.Sleep(time.Second * 2)
	cmd := exec.Command("telnet", "127.0.0.1", "8081")
	w, _ := cmd.StdinPipe()
	r, _ := cmd.StdoutPipe()
	defer func() {
		err := w.Close()
		if err != nil {
			log.Println("Failed to close writer!" + err.Error())
		}
	}()
	defer func() {
		err := r.Close()
		if err != nil {
			log.Println("Failed to close reader!" + err.Error())
		}
	}()
	var reader = bufio.NewReader(r)
	var writer = bufio.NewWriter(w)

	if err := cmd.Start(); err != nil {
		println("Failure: " + err.Error())
	}

	var readType = func() messageType {

		str, _ := reader.ReadString('}')
		str = regex.FindString(str)
		if len(str) == 0 {
			log.Fatal("EOF without }")
		}

		i, err := strconv.ParseUint(str[1:len(str)-1], 10, 64)
		if err != nil {
			log.Fatal("Failed to parse unsigned int from message!" + err.Error())
		}
		return messageType(i)
	}

	var readTypeIgnoreMessages = func() messageType {
		i := ChatMessage
		for {
			i = readType()
			if i != ChatMessage {
				return i
			}
		}
	}

	var writeFlush = func(m string) {
		_, err := writer.WriteString(m + "\n")
		if err != nil {
			log.Println("Failed to write during test!" + err.Error())
		}
		err = writer.Flush()
		if err != nil {
			log.Println("Failed to flush during test!" + err.Error())
		}
	}

	_, _ = reader.ReadString('\n')
	_, _ = reader.ReadString('\n')
	_, _ = reader.ReadString('\n')

	if readType() != NamePrompt {
		log.Fatal("Expected name prompt.")
	}

	writeFlush(name)

	for {
		time.Sleep(time.Second * time.Duration(rand.Int()%2+1))
		if readTypeIgnoreMessages() != ConsolePrompt {
			log.Print("Error! expected console prompt.")
		}
		writeFlush(randomUserMessage())
	}

}

func randomUserMessage() string {
	return randomMessages[rand.Int()%len(randomMessages)]
}

func main() {
	rand.Seed(time.Now().UnixNano())

	f, err := os.OpenFile("./test/test_logger.txt", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		fmt.Println("Failed to get logger file:" + err.Error())
	} else {
		defer func() {
			err := f.Close()
			if err != nil {
				log.Println("Failed to close writer!" + err.Error())
			}
		}()

	}

	log.SetFlags(log.Lshortfile)

	for i := 0; i < 5; i++ {
		go userTestLoop(userNames[i])
	}

	startServer()

}
