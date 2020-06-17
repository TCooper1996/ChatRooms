package main

import (
	"bufio"
	"log"
	"math/rand"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

type userTester struct {
	name   string
	reader *bufio.Reader
	writer *bufio.Writer
	proc   *exec.Cmd
	close  func()
}

var roomCountTest uint32 = 1
var roomCountTestMutex *sync.Mutex

func newUserTester(name string) *userTester {
	cmd := exec.Command("telnet", "127.0.0.1", "8081")
	w, _ := cmd.StdinPipe()
	r, _ := cmd.StdoutPipe()

	var reader = bufio.NewReader(r)
	var writer = bufio.NewWriter(w)

	var closePipes = func() {

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
	}
	return &userTester{name: name, reader: reader, writer: writer, proc: cmd, close: closePipes}
}

func (u *userTester) manageUserTester() {
	defer u.close()
	if err := u.proc.Start(); err != nil {
		println("Failure: " + err.Error())
	}

	_, _ = u.reader.ReadString('\n')
	_, _ = u.reader.ReadString('\n')
	_, _ = u.reader.ReadString('\n')

	if u.readType() != NamePrompt {
		log.Fatal("Expected name prompt.")
	}

	u.writeFlush(u.name)

	commands := []func(){
		func() {
			u.testRoomCreation()
		},
		func() {
			u.sendRandomMessage()
		},
	}

	for {
		time.Sleep(time.Second * time.Duration(rand.Int()%2+1))
		if u.readTypeIgnoreMessages() != ConsolePrompt {
			log.Print("Error! expected console prompt.")
		}
		u.writeFlush(randomUserMessage())

		commands[rand.Int()%len(commands)]()
	}

}

func (u *userTester) readType() messageType {
	str, _ := u.reader.ReadString('}')
	match := regex.FindString(str)

	if len(match) == 0 {
		log.Fatalln("EOF without } while reading message type. Received " + str)
	}

	i, err := strconv.ParseUint(match[1:len(match)-1], 10, 64)
	if err != nil {
		log.Fatalln("Failed to parse unsigned int from message!" + err.Error() + str)
	}

	return messageType(i)
}

func (u *userTester) readTypeIgnoreMessages() messageType {
	i := ChatMessage
	for {
		i = u.readType()
		if i != ChatMessage {
			return i
		}
	}
}

func (u *userTester) readTypeFilter(a ...messageType) messageType {
	var i messageType
	//todo: This is the worst code I've ever written.
	for {
		i = u.readType()
		for x, _ := range a {
			if i == a[x] {
				break
			}
			if x == (len(a) - 1) {
				return i
			}
		}
	}
}

func (u *userTester) writeFlush(m string) {
	_, err := u.writer.WriteString(m + "\n")
	if err != nil {
		log.Println("Failed to write during test!" + err.Error())
	}
	err = u.writer.Flush()
	if err != nil {
		log.Println("Failed to flush during test!" + err.Error())
	}
	_, _ = u.reader.ReadString('\r')
}

func (u *userTester) testRoomCreation() {
	log.Println("Testing room creation...")
	var roomName = u.name + "'sRoom"
	roomCountTestMutex.Lock()
	u.writeFlush("/create " + roomName)

	var t = u.readTypeFilter(ConsolePrompt, ChatMessage)
	switch t {
	case RoomCreated:
		if roomCountTest > roomLimit {
			log.Fatal("Room successfully created, but it should have failed, as the room limit was breached.")
		}
		roomCountTest++
		log.Println("Room created. Attempting to switch.")
		u.writeFlush("/switch " + roomName)
		if m := u.readTypeFilter(ChatMessage, ConsolePrompt); m != RoomChanged {
			log.Fatalln("Error: Expected RoomChanged messageType, got " + m.toString())
		}
		log.Println("Successfully changed rooms.")
	case RoomLimitReachedError:
		if roomCountTest != roomLimit {
			log.Fatalf("Room failed to create due to limit reached, but the current room count is %d, not %d!\n", roomCountTest, roomLimit)
		}
		log.Println("Room creation safely aborted, limit reached.")

	case RoomAlreadyExistsError:
		log.Println("Room creation safely aborted, room already exists.")

	default:
		log.Fatal("Room creation failed! Unexpected message type: " + t.toString())
	}
	roomCountTestMutex.Unlock()

}

func (u *userTester) sendRandomMessage() {
	u.writeFlush(randomUserMessage())
}
