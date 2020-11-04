# Team Cymru Telnet Chat Server

## Overview

This is a telnet based chat server that allows users to connect and send and receive text based communication. It is highly concurrent, implementing go routines, channels, mutexes, and sync Maps for the inter-thread communication.

### Telnet Server Procedure

When the application is started the configuration information is read in, the database is connected, the HTTP server is started in it's own thread and finally the telnet server is started.

The telnet server listens for new connections and once a connection is made a go routine to handle the client is started. This routine spawns a child routine which acts as the listener. The listener immediately waits for the username to be input. The username MUST BE UNIQUE. Once entered the user can begin to send and receive chat messages.

Once a user enters text and hits enter the application checks if the text matches any of the commands. If not then the message is broadcast. The broadcast message is placed into a sync Map which uses the name of each client as the key and the User struct as the value. Each listener checks for their own name as the key and deletes the key once received. There are sync Maps for broadcast, channel, and private messages. 

In order to keep the application performant channels are used to wait until a message has been sent. A controller struct is used that contains a start, done, and continueLoop channel attribute. The threadController() function handles the channels on the sender side. The process for that function is thus: place the number of connected clients represented as integers onto the start channel. Each listener will start receiving their own respective messages. Once finished each listener will send a done message and then wait to receive the continue loop message. The continue loop message WILL ONLY BE SENT once all listeners have sent their done message. REPEAT.

Finally, when a user connects the application checks the current number of clients, defined in the config file. If the number of clients has already been reached then that user is refused.

NOTE: The telnet server will only accept 1024 bytes per entry per session. 

## Features

### Logging

I decided to use the default logging application for simplicity. Upon starting the application the config package reads for any flags and then initializes the logging file. The logging file will use the name provided in the config file and appends the date. It stores all messages, errors, and any additional information like when a user connects or disconnects. It also logs to stdOut when a user connects, enters chat, and disconnects. Example logging messages are below.

    - ChatServer_1604353916 2020/11/02 BROADCAST MESSAGE - shanna Nov  2 16:52:26#: hello all
    - ChatServer_1604353916 2020/11/02 CHANNEL: 1 - MESSAGE - Channel: 1 Nov  2 16:54:19#: hello all from web
    - ChatServer_1604353760 2020/11/02 PRIVATE - RECIPIENT: stuart - MESSAGE - Private Message: andrew Nov  2 16:50:14#: hey stuart

### Config file

The config file is a json file that uses the below parameters. By default the file lives in the config directory but can be changed using the -file flag when starting the application.

- telnetPort
- httpPort
- maxClients
- logFile
- dialect
- connectionString

## Additional Features

### Commands
- /showusers: displays all connected users
- /quit: exits the chat application
- /exit: exits the chat application
- /ignore <user>: ignores all broadcast messages from a specific user
- /unignore: removes all ignored users from the ignore list
- /subscribe <channel number>: subscribe to channel and listen for messages sent to that channel
- /unsubscribe: stop subscribing to all channels
- /help: displays all commands

### API

I decided on using the included golang http package instead of any other packages like GIN or Gorilla Mux. The HTTP Server can accept connections for GET and POST requests to /chat. By default any GET request will receive a maximum of 100 messages from the database. This can be modified using the `limit` query string parameter. POST requests are used to send messages, either broadcast or channel messages. This is defined in the body of the message.

     - GET request parameters: "id", "user", "channel", "message", "recipient", "message_type", "limit"
     - POST request parameters: "message", "channel"

### Database

The database addition uses the Golang GORM package. This has default dialects for Postgres, Mysql, SQL Server, and SQL Lite. The config.json file contains configuration for the dialect and connection string. The GORM package has an AutoMigrate method which will auto create the table and columns. The database is used to store messages which are then retrieved using HTTP GET requests

**Postgresql**
- "dialect": "postgres",
- "connectionString": "postgres:\/\/postgres:password@localhost/dbname?sslmode=disable"

**Mysql/MariaDB**
- "dialect": "mysql",
- "connectionString": "root:password@tcp(localhost:3306)/dbname"


## 3rd Party Packages

### GORM

GORM is used for connecting to and interacting with the database.

## HOW TO RUN

From the root directory of the application type in `go run main.go`. This will kick off the application.

Default ports for telnet and HTTP are 23 and 80 respectively.

The default database is Postgresql. And uses the following connection string: `postgres://postgres:password@localhost/telnet?sslmode=disable`
You can change the credentials and database to anything you like.

If you intend to use another database application like Mysql then you will need to change the dialect import in db.go to `_ "github.com/jinzhu/gorm/dialects/mysql"` and run `go mod vendor`. See above for connection string for Mysql.

If intend to move the config file the please provide location using `go run main.go -file /file/config.json`.

## Known Bugs
- if there is data that hasn't been returned on screen and new data is sent from another user then you won't be able to delete that previous data but it will be sent on next return
    * appears to be related to the client. since the client only sends data on enter when a new message is received then the client still has the message waiting to return but the server has not received it yet.
    * to workaround just continue entering your message and hit enter OR hit enter immediately after receiving the chat message
- after subscribing the first channel message isn't received until a channel message is sent from the subscriber