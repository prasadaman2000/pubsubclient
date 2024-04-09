package pubsublib

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Message struct {
	topic   string
	message []byte
}

func (m *Message) GetTopic() string {
	return m.topic
}

func (m *Message) GetMessageString() string {
	return string(m.message)
}

func (m *Message) GetMessageRaw() []byte {
	return m.message
}

type PubSubClient struct {
	listenIp    string
	listenPort  int
	MessageChan chan *Message
}

func NewPubSubClient(listenIp string, listenPort int) *PubSubClient {
	client := &PubSubClient{
		listenIp:    listenIp,
		listenPort:  listenPort,
		MessageChan: make(chan *Message),
	}
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		topic := vals.Get("topic")
		if topic == "" {
			w.WriteHeader(400)
			w.Write([]byte("no topic found"))
			return
		}
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("could not get request body %v", err)))
			return
		}
		message := &Message{
			topic:   topic,
			message: body,
		}
		client.MessageChan <- message
	})
	go http.ListenAndServe(":"+strconv.Itoa(listenPort), serverMux)
	return client
}

func (p *PubSubClient) urlEncode() string {
	return fmt.Sprintf("peerIp=%s&peerPort=%d", p.listenIp, p.listenPort)
}

func (p *PubSubClient) Subscribe(topic string, serverIp string, serverPort int) error {
	serverUrl := fmt.Sprintf("http://%s:%d/subscribe?topic=%s&%s", serverIp, serverPort, topic, p.urlEncode())
	resp, err := http.Get(serverUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Subscribe failed with error: %s", string(body))
	}
	fmt.Printf("subscribed! %v %d\n", string(body), resp.StatusCode)
	return nil
}

func (p *PubSubClient) Publish(topic string, msg []byte, serverIp string, serverPort int) error {
	serverUrl := fmt.Sprintf("http://%s:%d/publish?topic=%s&%s", serverIp, serverPort, topic, p.urlEncode())
	resp, err := http.Post(serverUrl, "text", bytes.NewBuffer(msg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Publish failed with error: %s", string(body))
	}
	return nil
}
