package server

/*
	OVERVIEW: Server is the primary package in the application. This performs the majority of the work of the telnet server.
	It uses a go 'net' listener, the port for this is defined in config.json.
	There are a total of three files in this package. 1. server.go, 2. chat.go, 3. commands.go
	server.go is intended to handle communication for the server.
	chat.go is intended to handle communication for the client.
	commands.go is intended to handle commands from the client.
	The telnet port
*/

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"team-cymru-telnet/config"
	"time"
)

//Channel used to block the go routines from iterating infinitely without any control and consuming excessive CPU cycles
type controller struct {
	//start channel starts the chatListener loop - this is sent by any of the Send methods
	start chan int
	//done channel informs the sending method that one of the listening threads has finished
	done chan int
	//continue loop informs all of the listening threads that they can continue to the beginning again
	continueLoop chan int
}

var loopController controller = controller{
	start:        make(chan int),
	done:         make(chan int),
	continueLoop: make(chan int),
}

func Start() {
	service := fmt.Sprintf(":%s", config.Cfg.TelnetPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	config.CheckError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	config.CheckError(err)

	numClients := 0
	config.Logs().Info("chat server has started")
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		addr := conn.RemoteAddr()
		//num clients limits the number of clients BEFORE a user is added to connectedClients
		numClients++

		if numClients > config.Cfg.MaxClients {
			conn.Write([]byte("Connection Refused. Too many current clients. Please try again later.\r\n"))
			config.Logs().Info(fmt.Sprintf("new client %s attempted to connect. connection refused. too many current clients", addr.String()))
			conn.Close()
			continue
		}

		config.Logs().Info(fmt.Sprintf("new client %s has connected", addr.String()))

		go handleClient(conn, &numClients)
	}
}

func handleClient(conn net.Conn, numClients *int) {
	// close connection on exit
	defer conn.Close()

	user := User{}
	mutex := sync.Mutex{}
	ignoreUserMap := make(map[string]bool)
	buildIgnoreMap(ignoreUserMap)
	nameChan := make(chan string, 1)
	go messageListener(nameChan, ignoreUserMap, &user, conn)

	buf := [1024]byte{}
	line := ""

	conn.Write([]byte("Please enter name\r\n#:"))
	for {
		user.TimeStamp = time.Now().Local().Format(time.Stamp)
		n, err := conn.Read(buf[0:])
		if err != nil {
			config.CheckError(err)
			return
		}

		//gather the input
		buff := string(buf[0:n])
		line += buff

		//write the output to the screen
		conn.Write(buf[n:])
		//reset buffer
		buf = [1024]byte{}

		//the MacOS in the default telnet client sends the entire string upon return. While Windows sends each character.
		//the below
		if strings.Contains(buff, "\r\n") { //once return is entered then send the message
			line = newLineTrim(line)
			if user.Name == "" {
				name := line //look for the name first and set it
				line = ""
				if strings.Contains(name, " ") {
					config.Logs().Error("Name cannot contain spaces.")
					conn.Write([]byte("Name cannot contain spaces\r\n#:"))
					conn.Write([]byte("Please enter name\r\n#:"))
					continue
				}

				if !AddUniqueClient(name) {
					config.Logs().Error(fmt.Sprintf("User, %s, already exists in chat", name))
					conn.Write([]byte("User already exists in chat\r\n#:"))
					conn.Write([]byte("Please enter name\r\n#:"))
					continue
				}
				nameChan <- name
				close(nameChan)
				continue
			}
			switch {
			case Exit(line):
				conn.Write([]byte("closing connection"))
				config.Logs().Info(fmt.Sprintf("closing client %s", user.Name))

				//Lock and unlock numClients since we could potentially have one connection exit simultaneously we want to
				//lock this variable from being modified by multiple threads concurrently
				mutex.Lock()
				*numClients--
				removeClient(user.Name)
				mutex.Unlock()
				return
			case line == showUsers:
				displayUsers(conn)
			case ignore.MatchString(line):
				//protect the ignoreUserMap variable from potentially reading and writing to ignoreUserMap in the listener
				mutex.Lock()
				updateIgnoreMap(ignoreUserMap, line, 1, conn)
				mutex.Unlock()
			case line == unIgnore:
				mutex.Lock()
				//protect ignoreUserMap
				resetIgnoreUserMap(ignoreUserMap)
				mutex.Unlock()
				conn.Write([]byte("now allowing messages from all users"))
			case channel.MatchString(line):
				updateUserWithChannel(&user, line)
				SendToChannel(user)
				line = ""
				continue
			case pm.MatchString(line):
				updateUserPM(&user, line)
				sendPM(user)
			case subscribe.MatchString(line):
				addChannel(user.Name, line, 1, conn)
			case line == unsubscribe:
				unsubscribeChannels(user)
				conn.Write([]byte("ceased subscribing to all channels"))
			case line == help:
				displayHelp(conn)
			default:
				// fmt.Println(line)
				if line != "" {
					user.Message = line
					SendBroadcast(user)
					line = ""
					continue
				}

			}
			line = ""
			conn.Write([]byte("\r\n" + user.Name + "#: "))
		}
	}
}

func newLineTrim(line string) string {
	line = strings.Trim(line, "\r\n")
	line = strings.Trim(line, "\n")
	line = strings.Trim(line, "\r")
	return line
}

func pmListener(name string, conn net.Conn) {
	if value, ok := privateMessage.Load(name); ok {
		conn.Write([]byte("\r\n" + value.(string) + "\r\n" + name + "#: "))
	}
}

func channelListener(user User, conn net.Conn) {
	if user.Channel == 0 {
		return
	}

	if value, ok := channelMessage.Load(user.Channel); ok {
		for _, client := range connectedClients {
			if user.Name == client.name {
				for _, num := range client.channels {
					if num == user.Channel {
						conn.Write([]byte("\r\n" + value.(string) + "\r\n" + user.Name + "#: "))
					}
				}
				return
			}
		}
	}
}

func broadCastListener(name string, ignoreUserMap map[string]bool, conn net.Conn) {
	if value, ok := broadCastMessage.Load(name); ok {
		sender := value.(User)
		//if the sender of the msg is FALSE in the ignore list then continue and don't print to screen
		if value := ignoreUserMap[sender.Name]; !value {
			return
		}
		conn.Write([]byte("\r\n" + sender.Message + "\r\n" + name + "#: "))
		//delete key in order to prevent weird additional messaging
		broadCastMessage.Delete(name)
	}
}

//function to first grab the name of the user and once received begin listening for chat messages
func messageListener(nameChan <-chan string, ignoreUserMap map[string]bool, user *User, conn net.Conn) {
	//if name is empty still then we block until a name is received from the parent thread
	log.Println("waiting for name...")
	for user.Name = range nameChan {
	}
	config.Logs().Info(fmt.Sprintf("user %s has entered chat...", user.Name))
	conn.Write([]byte("\r\n" + user.Name + "#: "))

	chatListener(ignoreUserMap, user, conn)
}

//listens for broadcast chat messages
func chatListener(ignoreUserMap map[string]bool, user *User, conn net.Conn) {
	keyCount := keyCounter(ignoreUserMap)
	//block until a value can be discarded from the channel
	for range loopController.start {
		//client has left chat or new client has joined
		if keyCount != len(connectedClients) {
			buildIgnoreMap(ignoreUserMap)
		}

		broadCastListener(user.Name, ignoreUserMap, conn)
		channelListener(*user, conn)
		pmListener(user.Name, conn)
		keyCount = keyCounter(ignoreUserMap)

		loopController.done <- 1
		<-loopController.continueLoop
	}
}

func buildIgnoreMap(ignoreMap map[string]bool) {
	for _, client := range connectedClients {
		if ignoreMap[client.name] {
			continue
		}
		ignoreMap[client.name] = true
	}
}

//modify the ignore map if the number of connected clients has changed
func keyCounter(ignoreUserMap map[string]bool) int {
	count := 0
	for count < len(ignoreUserMap) {
		count++
	}
	return count
}
