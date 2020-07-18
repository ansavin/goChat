package main

import(
	"os"
	"fmt"
	"net"
	"math/rand"
	"time"
	"bufio"
	"strings"
)

type User struct {
	Id int
	Username string
	Ch chan string
	Status int
}

type Thread []User

const(
	Offilne = iota
	Online
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
			fmt.Println("Please provide a port number!")
			return
	}

	port := ":" + arguments[1]
	l, err := net.Listen("tcp4", port)
	if err != nil {
			fmt.Println(err)
			return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	var t Thread
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		t = append(t, handleConnection(c, &t))
		fmt.Println("Got new connection!")
	}
}

func handleConnection(c net.Conn, input *Thread) User {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	ch := make(chan string, 5)

	id := rand.Int()

	go communicate(c, input, ch, id)

	return User{
		Id: id,
		Username: "user" + string(id),
		Ch: ch,
		Status: Online,
	}
}

func communicate(c net.Conn, input *Thread, out chan string, currentUserId int) {
	textMsgChan := make(chan string)
	go waitForTextInput(c, currentUserId, textMsgChan)
	for {
		select {
		case incomingText, ok := <- textMsgChan: 
		    fmt.Println("Wait for text input...")
			if !ok {
				fmt.Printf("User %d chan is closed!\n", currentUserId)
				return
			}
			
			if incomingText == `\exit` {
				fmt.Printf("User %d left chat!\n", currentUserId)
				return
			}

			broadcast(incomingText, input, currentUserId)
		case incomingMsg := <- out:
			fmt.Printf("User %d has a new message!\n", currentUserId)
			c.Write([]byte(incomingMsg))
		}
	}
}

func waitForTextInput(c net.Conn, currentUserId int, out chan string) {
	for {
		incomingText, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			close(out)
			return
		}
		fmt.Printf("User %d typed msg %s\n", currentUserId, incomingText)
		incomingText = strings.TrimSpace(incomingText) + "\n"
		out <- string(incomingText)
	}
}

func broadcast(msg string, input *Thread, currentUserId int) {
	fmt.Println("Enter broadcast...")
	for _,user := range *input {
		if user.Id == currentUserId{
			fmt.Println("Skip own chan!")
			continue
		}
		fmt.Printf("Write to user %d chan...\n", user.Id)
		user.Ch <- msg
		fmt.Println("Done!")
	}
	fmt.Println("Quit broadcast!")
}