// Package pubsubclient further abstracts pubsub
package main

import (
	"flag"
	"fmt"
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
	address := flag.String("address", "0.0.0.0", "--address defines the address to bind the server to.")
	port := flag.Int("port", 8099, "--port defines the port to bind the server to.")

	pubsubServerAddress := flag.String("pubsub_addr", "0.0.0.0", "--pubsub_addr defines the address the base pubsub server is running on.")
	pubsubServerPort := flag.Int("pubsub_port", 8080, "--port defines the port the base pubsub server is running on.")

	flag.Parse()

	pubsubclientAddr := fmt.Sprintf("%s:%d", *address, *port)

	p := NewPopulation(*pubsubServerAddress, *pubsubServerPort, *address)
	http.HandleFunc("/clientConnect", p.ClientConnectEntry)
	http.HandleFunc("/subscribe", p.SubscribeEntry)
	http.HandleFunc("/publish", p.PublishEntry)
	http.HandleFunc("/poll", p.PollMessagesEntry)

	http.ListenAndServe(pubsubclientAddr, nil)
}
