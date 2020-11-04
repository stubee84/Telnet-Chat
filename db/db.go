package db

/*
	OVERVIEW: database model for the backend connection to the database. This uses the database connection information stored in config.json.
	It uses the GORM package to connect to and communicate with the database.
	GORM has some minor migrations built into the package so if you add a column to a table struct then the actual table will be updated.
*/

import (
	"sync"
	"team-cymru-telnet/config"
	"team-cymru-telnet/models/chat"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var mutex sync.Mutex = sync.Mutex{}

//Will auto update any table that gorm has a connection to
func autoMigrate() {
	ch := chat.Chat{}
	DB.Conn.Debug().AutoMigrate(&ch)
}

func Connect() {
	var err error
	//Lock the read and writing to the DB variable to prevent mutiple DB connections
	mutex.Lock()
	if !DB.Connected {
		DB.Conn, err = gorm.Open(config.Cfg.Dialect, config.Cfg.ConnectionString)
		if err != nil {
			config.Logs().Fatal(err.Error())
		}

		DB.Connected = true
		autoMigrate()
	}
	mutex.Unlock()
}

type DBConn struct {
	Conn      *gorm.DB
	Connected bool
}

var DB *DBConn = &DBConn{
	Conn:      nil,
	Connected: false,
}
