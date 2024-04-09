// Package pubsubclient further abstracts pubsub
package main

import (
	"net/http"
)

const SERVER_IP = "localhost"
const SERVER_PORT = 8080

/*
	/clientConnect - takes string username and string password in URL params
	/subscribe - takes string username, string password, and string topic in URL params
	/publish - takes string username, string password, and string topic in URL params and arbitrary bytes in req body
	/poll - takes string username, string password, returns JSON serialized array of QueuedMessages (byte array is a literal array of bytes)
*/
func main() {
	p := NewPopulation(SERVER_IP, SERVER_PORT)
	http.HandleFunc("/clientConnect", p.ClientConnectEntry)
	http.HandleFunc("/subscribe", p.SubscribeEntry)
	http.HandleFunc("/publish", p.PublishEntry)
	http.HandleFunc("/poll", p.PollMessagesEntry)

	http.ListenAndServe(":8099", nil)
}
