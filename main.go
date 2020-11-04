package main

/*
	OVERVIEW: Main packages initlizes config, db, and starts the HTTP and Telnet servers.
*/

import (
	"team-cymru-telnet/api"
	"team-cymru-telnet/config"
	"team-cymru-telnet/db"
	"team-cymru-telnet/server"
)

func main() {
	config.Init()
	db.Connect()

	api.Start()
	server.Start()
}
