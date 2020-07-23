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
	Name string
	Ch chan string
	Status int
}

type Thread map[int]*User

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

	t := make(Thread)
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		newUser := handleConnection(c, &t)
		if newUser.Id == 0 {
			continue
		}
		t[newUser.Id] = newUser
		fmt.Println("Got new connection!")
	}
}

func handleConnection(c net.Conn, input *Thread) *User {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	ch := make(chan string, 5)

	id := rand.Int()

	c.Write([]byte("Enter your username:\n"))

	go communicate(c, input, ch, id)

	return &User{
		Id: id,
		Name: "anonimus",
		Ch: ch,
		Status: Online,
	}
}

func communicate(c net.Conn, input *Thread, out chan string, currentUserId int) {
	defer c.Close()

	// We assume that the first thing user should write is his name
	renameInProgress := true

	textMsgChan := make(chan string)

	go waitForTextInput(c, input, currentUserId, textMsgChan)

	for {
		select {
		case incomingText, ok := <- textMsgChan: 
			fmt.Println("Wait for text input...")
			if !ok {
				fmt.Printf("User %d chan is closed!\n", currentUserId)
				return
			}

			if incomingText == "\\exit" {
				fmt.Printf("User %d left chat!\n", currentUserId)
				broadcast(
					fmt.Sprintf("User %s left the chat\n", (*input)[currentUserId].Name),
					input,
					currentUserId)
				close(out)
				delete(*input, currentUserId)
				return
			}
	
			if incomingText == "\\users" {
				fmt.Printf("User %d wants to list chat users!\n", currentUserId)
				msg := fmt.Sprintf("Now we have %d users online:\n", len(*input))
				for _,user := range *input{
					msg += fmt.Sprintf("%s\n",user.Name)
				}
				c.Write([]byte(msg))
				continue
			}
	
			if incomingText == "\\rename" {
				fmt.Printf("User %d wants to change name!\n", currentUserId)
				c.Write([]byte("Enter new name:\n"))
				renameInProgress = true
				continue
			}
			
			if renameInProgress {
				fmt.Printf("User %d entered new name!\n", currentUserId)
	
				oldName := (*input)[currentUserId].Name
	
				(*(*input)[currentUserId]).Name = incomingText
	
				fmt.Printf("User %d changed name successfully!\n", currentUserId)
	
				broadcast(
					fmt.Sprintf("User %s change name to %s\n",
						oldName,
						(*input)[currentUserId].Name),
					input,
					currentUserId)
				renameInProgress = false
				continue
			}

			if !renameInProgress {
				incomingText = incomingText + "\n"

				broadcast(incomingText, input, currentUserId)
			}

		case incomingMsg := <- out:
			fmt.Printf("User %d has a new message!\n", currentUserId)
			c.Write([]byte(incomingMsg))
		}
	}
}

func waitForTextInput(c net.Conn, input *Thread, currentUserId int, out chan string) {
	for {
		incomingText, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			close(out)
			return
		}

		fmt.Printf("User %d typed msg %s\n", currentUserId, incomingText)

		out <- strings.TrimSpace(string(incomingText))
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

		user.Ch <- fmt.Sprintf("[%s]: %s", (*input)[currentUserId].Name, msg)

		fmt.Println("Done!")
	}
	fmt.Println("Quit broadcast!")
}