package server

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"team-cymru-telnet/config"
)

func Exit(line string) bool {
	if strings.ToLower(line) == exit || strings.ToLower(line) == quit {
		return true
	}
	return false
}

func displayHelp(conn net.Conn) {
	commands := []string{"/exit", "/quit", "/showusers", "/ignore <user>", "/unignore", "/channel <channel number> <message>", "/pm <user> <message>",
		"/subscribe <channel number>", "/unsubscribe", "/help"}
	// commands := map[string]string{
	// 	"/exit":                               "quit the chat application\r\n",
	// 	"/quit":                               "quit the chat application\r\n",
	// 	"/showusers":                          "list all connected users\r\n",
	// 	"/ignore <user>":                      "ignore all messages from user\r\n",
	// 	"/unignore":                           "remove all users from ignore list\r\n",
	// 	"/channel <channel number> <message>": "send message to channel\r\n",
	// 	"/pm <user> <message>":                "send private message to user\r\n",
	// 	"/subscribe <channel number>":         "subscribe to channel\r\n",
	// 	"/unsubscribe":                        "stop channel subscription\r\n",
	// 	"/help":                               "displays this information\r\n",
	// }
	helpMsg := ""
	for _, value := range commands {
		switch value {
		case "/exit":
			helpMsg = fmt.Sprintf("%s: quit the chat application\r\n", value)
		case "/quit":
			helpMsg = fmt.Sprintf("%s: quit the chat application\r\n", value)
		case "/showusers":
			helpMsg = fmt.Sprintf("%s: list all connected users\r\n", value)
		case "/ignore <user>":
			helpMsg = fmt.Sprintf("%s: ignore all messages from user\r\n", value)
		case "/unignore":
			helpMsg = fmt.Sprintf("%s: remove all users from ignore list\r\n", value)
		case "/channel <channel number> <message>":
			helpMsg = fmt.Sprintf("%s: send message to channel\r\n", value)
		case "/pm <user> <message>":
			helpMsg = fmt.Sprintf("%s: send private message to user\r\n", value)
		case "/subscribe <channel number>":
			helpMsg = fmt.Sprintf("%s: subscribe to channel\r\n", value)
		case "/unsubscribe":
			helpMsg = fmt.Sprintf("%s: stop channel subscription\r\n", value)
		case "/help":
			helpMsg = fmt.Sprintf("%s: displays this information\r\n", value)
		}
		conn.Write([]byte(helpMsg))
	}
	conn.Write([]byte("\r\n"))
}

func displayUsers(conn net.Conn) {
	for _, client := range connectedClients {
		if client.name != "web" {
			conn.Write([]byte(client.name))
			conn.Write([]byte("\r\n"))
		}
	}
}

func updateIgnoreMap(ignoreMap map[string]bool, line string, index int, conn net.Conn) {
	for i, str := range ignore.FindStringSubmatch(line) {
		if i == index {
			ignoreMap[str] = false
			conn.Write([]byte(fmt.Sprintf("now ignoring user %s\n", str)))
			return
		}
	}
}

func resetIgnoreUserMap(ignoreMap map[string]bool) {
	for key := range ignoreMap {
		ignoreMap[key] = true
	}
}

func unsubscribeChannels(user User) {
	for i := range connectedClients {
		if user.Name == connectedClients[i].name {
			connectedClients[i].channels = []int{}
		}
	}
}

//update the User struct for the below attributes which will then be used for a listener
func updateUserPM(user *User, line string) {
	for i, str := range pm.FindStringSubmatch(line) {
		if i == 1 {
			user.Recipient = str
		} else if i == 2 {
			user.Message = str
		}
	}
}

func addChannel(name string, line string, lineIndex int, conn net.Conn) {
	channelNum := 0
	var err error
	for i, str := range subscribe.FindStringSubmatch(line) {
		if i == lineIndex {
			channelNum, err = strconv.Atoi(str)
			if err != nil {
				config.Logs().Error(fmt.Sprintf("failed to subscribe to channel %s. error: %v", str, err))
				conn.Write([]byte(fmt.Sprintf("could not subscribe to channel. %v", err)))
				return
			}
			break
		}
	}

	for i := range connectedClients {
		if connectedClients[i].name == name {
			//prevent adding duplicate channels to slice
			for _, num := range connectedClients[i].channels {
				if num == channelNum {
					conn.Write([]byte(fmt.Sprintf("channel %d has already been added", channelNum)))
					return
				}
			}
			connectedClients[i].channels = append(connectedClients[i].channels, channelNum)
			conn.Write([]byte(fmt.Sprintf("now subscribing to channel %d\n", channelNum)))
			return
		}
	}
}

//updates the User struct with the channel number
func updateUserWithChannel(user *User, line string) {
	var err error
	for i, str := range channel.FindStringSubmatch(line) {
		if i == 1 {
			user.Channel, err = strconv.Atoi(str)
			if err != nil {
				config.Logs().Error(fmt.Sprintf("failed to send to channel %s. error: %v", str, err))
			}
		} else if i == 2 {
			user.Message = str
		}
	}
}

var exit string = "/exit"
var quit string = "/quit"
var showUsers string = "/showusers"
var ignore *regexp.Regexp = regexp.MustCompile("^/ignore ([a-z]+)$")
var unIgnore string = "/unignore"
var channel *regexp.Regexp = regexp.MustCompile("^/channel (\\d+) (.*)$")
var pm *regexp.Regexp = regexp.MustCompile("^/pm ([a-z]+) (.*)$")
var subscribe *regexp.Regexp = regexp.MustCompile("^/subscribe (\\d+)$")
var unsubscribe string = "/unsubscribe"
var help string = "/help"
