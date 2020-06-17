package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"sync"
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

var regex, _ = regexp.Compile("{[0-9]+}")

func randomUserMessage() string {
	return randomMessages[rand.Int()%len(randomMessages)]
}

func main() {
	rand.Seed(time.Now().UnixNano())
	roomCountTestMutex = &sync.Mutex{}

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
		go newUserTester(userNames[i]).manageUserTester()
	}

	startServer()

}
