# PubsubClient
## A poorly-named message-queue-based wrapper on top of pubsublib

PubsubClient uses message queues built on top of my raw pubsub library to help support applications that can't have long running HTTP servers to listen for published messages.

### PubsubClient starts an HTTP server on port 8099 that supports the following routes:
* `/clientConnect` registers a user as someone who can receive messages on this server
    * Required URL params
        * `username`: string representing the username of who is registering
        * `password`: string representing a password that will be used to authenticate the user for all further operations
* `/subscribe` subscribes a user with a given username to a topic
    * Required URL params
        * `username`: string representing the username of who is subscribing
        * `password`: string representing the password a user registered with
        * `topic`: string representing the topic the user is trying to subscribe to
* `/publish` publishes a message to a particular topic
    * Required URL params
        * `username`: string representing the username of who is publishing
        * `password`: string representing the password a user registered with
        * `topic`: string representing the topic the user is trying to publish to
    * Body
        * Can be anything really, this abstraction accepts any bytes as the published message
* `/poll` polls all messages received by the user. Messages can only be polled once and the queue is cleared upon polling.
    * Required URL params
        * `username`: string representing the username of who is polling
        * `password`: string representing the password a user registered with
    * Returns
        * JSON encoded array of messages. Fields are in the QueuedMessage type in `psclienthelpers.go`


### Notes:
* Run this [pubsub](https://github.com/prasadaman2000/pubsub) server. This will start a compatible pubsub server on port 8080, which is required for the client to work
