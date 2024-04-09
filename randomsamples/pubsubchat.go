package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"pubsubclient/pubsublib"
	"sync"
	"time"
)

const ServerIp string = "localhost"
const ServerPort int = 8080

type UserMessage struct {
	Username string
	Time     time.Time
	Msg      string
}

func getMessages(mutex *sync.Mutex, ps *pubsublib.PubSubClient) {
	for {
		msgBytes := <-ps.MessageChan
		userMsg := &UserMessage{}
		err := json.Unmarshal(msgBytes.GetMessageRaw(), userMsg)
		if err != nil {
			fmt.Printf("could not unmarshal message %s\n", msgBytes.GetMessageString())
			continue
		}
		mutex.Lock()
		fmt.Printf("\r[%s] %s: %s", userMsg.Time.Local().Format("02 Jan 06 15:04 MST"), userMsg.Username, userMsg.Msg)
		fmt.Printf("enter a message: ")
		mutex.Unlock()
	}
}

func main() {
	portFlag := flag.Int("port", 8000, "--port defines the port to listen to messages on")
	usernameFlag := flag.String("username", "", "--username defines the username with which a user will appear to other users in the room")
	roomFlag := flag.String("room", "", "--room defines the name of the room the user will connect to")
	flag.Parse()

	client := pubsublib.NewPubSubClient("localhost", *portFlag)
	err := client.Subscribe(*roomFlag, ServerIp, ServerPort)
	if err != nil {
		fmt.Printf("could not subscribe to room %s: %v", *roomFlag, err)
		return
	}

	chatMutex := &sync.Mutex{}
	go getMessages(chatMutex, client)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("enter a message: ")
		str, err := reader.ReadString('\n')
		chatMutex.Lock()
		if err != nil {
			fmt.Printf("could not read from stdin: %v\n", err)
			chatMutex.Unlock()
			continue
		}
		userMsg := &UserMessage{
			Username: *usernameFlag,
			Msg:      str,
			Time:     time.Now(),
		}
		marshalledBytes, err := json.Marshal(userMsg)
		if err != nil {
			fmt.Printf("could not marshal: %v\n", err)
			chatMutex.Unlock()
			continue
		}
		client.Publish(*roomFlag, marshalledBytes, ServerIp, ServerPort)
		chatMutex.Unlock()
	}
}
