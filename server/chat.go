package server

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"team-cymru-telnet/config"
	"team-cymru-telnet/db"
	"team-cymru-telnet/models/chat"
)

//object representing the connected user
type User struct {
	Name      string
	Recipient string
	TimeStamp string
	Channel   int
	Message   string
}

//Send - create a unique channel per connected client. Using the connected client name as the key
func SendBroadcast(user User) bool {
	chat := &chat.Chat{
		User:        user.Name,
		MessageType: "broadcast",
		Message:     user.Message,
	}

	msg := fmt.Sprint(user.Name + " " + user.TimeStamp + "#: " + user.Message)
	user.Message = msg
	for _, client := range connectedClients {
		client.msg = msg
		broadCastMessage.Store(client.name, user)
	}

	if !threadController() {
		if user.Name == "web" {
			config.Logs().Info("could not receive message from web. no available listening clients")
		}
		return false
	}

	config.Logs().FileLogger.Printf("BROADCAST MESSAGE - %s", msg)
	db.DB.Conn.Save(chat)

	return true
}

func SendToChannel(user User) bool {
	chat := &chat.Chat{
		User:        user.Name,
		MessageType: "channel",
		Channel:     sql.NullInt64{Valid: true, Int64: int64(user.Channel)},
		Message:     user.Message,
	}
	msg := fmt.Sprint("Channel: " + strconv.Itoa(user.Channel) + " " + user.TimeStamp + "#: " + user.Message)
	channelMessage.Store(user.Channel, msg)

	if !threadController() {
		if user.Name == "web" {
			config.Logs().Info("could not receive message from web. no available listening clients")
		}
		return false
	}

	channelMessage.Delete(user.Channel)

	config.Logs().FileLogger.Printf("CHANNEL: %d - MESSAGE - %s", user.Channel, msg)
	db.DB.Conn.Save(chat)

	return true
}

func sendPM(user User) {
	chat := &chat.Chat{
		User:        user.Name,
		MessageType: "pm",
		PMRecipient: sql.NullString{Valid: true, String: user.Recipient},
		Message:     user.Message,
	}

	msg := fmt.Sprint("Private Message: " + user.Name + " " + user.TimeStamp + "#: " + user.Message)
	privateMessage.Store(user.Recipient, msg)

	threadController()

	privateMessage.Delete(user.Recipient)

	config.Logs().FileLogger.Printf("PRIVATE - RECIPIENT: %s - MESSAGE - %s", user.Recipient, msg)
	db.DB.Conn.Save(chat)
}

func threadController() bool {
	//if there is only 1 client then it is "web"
	if clientCount == 1 {
		return false
	}

	counter := 0
	for _, client := range connectedClients {
		//since web does not listen for any messages we do not add any values to the channel queue
		if client.name != "web" {
			loopController.start <- 1
		}
	}

	for counter < clientCount-1 {
		<-loopController.done
		counter++
	}

	for counter > 0 {
		counter--
		loopController.continueLoop <- 1
	}
	return true
}

func AddUniqueClient(name string) bool {
	for _, client := range connectedClients {
		if name == client.name {
			return false
		}
	}
	client := client{
		name:    name,
		sending: false,
		counter: 0,
	}
	connectedClients = append(connectedClients, client)
	clientCount++

	return true
}

func removeClient(name string) {
	index := 0
	for i, client := range connectedClients {
		if client.name == name {
			index = i
		}
	}
	connectedClients[len(connectedClients)-1], connectedClients[index] = connectedClients[index], connectedClients[len(connectedClients)-1]
	connectedClients = connectedClients[:len(connectedClients)-1]
	clientCount--
}

type client struct {
	name     string
	channels []int
	sending  bool
	counter  int
	msg      string
}

var broadCastMessage sync.Map = sync.Map{}
var channelMessage sync.Map = sync.Map{}
var privateMessage sync.Map = sync.Map{}
var connectedClients []client = []client{}
var clientCount int = 0
