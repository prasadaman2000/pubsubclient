package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pubsubclient/pubsublib"
	"strings"
	"sync"
	"time"
)

const LOCAL_ADDR = "localhost"

type JSONableSlice []byte

func (u JSONableSlice) MarshalJSON() ([]byte, error) {
	var result string
	if u == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", u)), ",")
	}
	return []byte(result), nil
}

type QueuedMessage struct {
	Msg  JSONableSlice
	Time time.Time
}

type Client struct {
	ip           string
	port         int
	username     string
	password     string
	messageQueue []*QueuedMessage // this addr will likely be constantly changing, don't read this directly
	lastOutgoing time.Time
	pubSubClient *pubsublib.PubSubClient
	lock         *sync.Mutex
}

func (c *Client) EnqueueMessages() {
	for {
		msg := <-c.pubSubClient.MessageChan
		message := &QueuedMessage{
			Msg:  msg.GetMessageRaw(),
			Time: time.Now().UTC(),
		}
		c.lock.Lock()
		c.messageQueue = append(c.messageQueue, message)
		c.lock.Unlock()
	}
}

func (c *Client) PollMessages() []*QueuedMessage {
	c.lock.Lock()
	msgs := c.messageQueue
	c.messageQueue = make([]*QueuedMessage, 0)
	c.lock.Unlock()
	return msgs
}

func (c *Client) Subscribe(topic string, serverIp string, serverPort int) error {
	return c.pubSubClient.Subscribe(topic, serverIp, serverPort)
}

func (c *Client) Publish(topic string, message []byte, serverIp string, serverPort int) error {
	err := c.pubSubClient.Publish(topic, message, serverIp, serverPort)
	if err != nil {
		return err
	}
	c.lastOutgoing = time.Now().UTC()
	return nil
}

type Population struct {
	clientMap  map[string]*Client
	serverIp   string
	serverPort int
}

func NewPopulation(serverIp string, serverPort int) *Population {
	return &Population{
		clientMap:  make(map[string]*Client),
		serverIp:   serverIp,
		serverPort: serverPort,
	}
}

var portNum int = 8100

// var listenIp string = "localhost"
var listenIp string = LOCAL_ADDR

func (p *Population) ClientConnect(username string, password string) error {
	if user, ok := p.clientMap[username]; ok {
		if user.password != password {
			return fmt.Errorf("user %s exists, tried to be recreated with wrong password", username)
		}
		fmt.Printf("User %s is already registered\n", username)
		return nil
	}
	psClient := pubsublib.NewPubSubClient(listenIp, portNum)
	client := &Client{
		ip:           listenIp,
		port:         portNum,
		password:     password,
		messageQueue: make([]*QueuedMessage, 0),
		lastOutgoing: time.Now(),
		pubSubClient: psClient,
		username:     username,
		lock:         &sync.Mutex{},
	}
	portNum++
	p.clientMap[username] = client
	go client.EnqueueMessages()
	return nil
}

func (p *Population) PollClientMessages(username string, password string) ([]*QueuedMessage, error) {
	if user, ok := p.clientMap[username]; ok {
		if user.password != password {
			return nil, fmt.Errorf("user %s either does not exist or bad password", username)
		}
		return user.PollMessages(), nil
	}
	return nil, fmt.Errorf("user %s either does not exist or bad password", username)
}

func (p *Population) Subscribe(username string, password string, topic string) error {
	if user, ok := p.clientMap[username]; ok {
		if user.password != password {
			return fmt.Errorf("user %s either does not exist or bad password", username)
		}
		return user.Subscribe(topic, p.serverIp, p.serverPort)
	}
	return fmt.Errorf("user %s either does not exist or bad password", username)
}

func (p *Population) Publish(username string, password string, topic string, message []byte) error {
	if user, ok := p.clientMap[username]; ok {
		if user.password != password {
			return fmt.Errorf("user %s either does not exist or bad password", username)
		}
		return user.Publish(topic, message, p.serverIp, p.serverPort)
	}
	return fmt.Errorf("user %s either does not exist or bad password", username)
}

func GetURLArg(r *http.Request, arg string) (string, error) {
	vals := r.URL.Query()
	var err error
	val := vals.Get(arg)
	if val == "" {
		err = fmt.Errorf("GetURLArg could not get argument %v", arg)
	}
	return val, err
}

func (p *Population) ClientConnectEntry(w http.ResponseWriter, r *http.Request) {
	username, err := GetURLArg(r, "username")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	password, err := GetURLArg(r, "password")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	err = p.ClientConnect(username, password)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("client %s connected", username)))
}

func (p *Population) SubscribeEntry(w http.ResponseWriter, r *http.Request) {
	username, err := GetURLArg(r, "username")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	password, err := GetURLArg(r, "password")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	topic, err := GetURLArg(r, "topic")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	err = p.Subscribe(username, password, topic)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("subscribed to %v", topic)))
}

func (p *Population) PublishEntry(w http.ResponseWriter, r *http.Request) {
	username, err := GetURLArg(r, "username")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	password, err := GetURLArg(r, "password")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	topic, err := GetURLArg(r, "topic")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf("could not get request body %v", err)))
		return
	}
	err = p.Publish(username, password, topic, body)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("published to %v", topic)))
}

func (p *Population) PollMessagesEntry(w http.ResponseWriter, r *http.Request) {
	username, err := GetURLArg(r, "username")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	password, err := GetURLArg(r, "password")
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	messages, err := p.PollClientMessages(username, password)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	marshaled, err := json.Marshal(messages)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(marshaled)
}
