package config

/*
	OVERVIEW: The config package receives the application configuration from config.json. Initializes the app flags, config file, and logging.
	The only command flag that can be used is -file. This specifies the file name and path for the configuration file.
	Logs() initializes the fLogger (file logger) variable and returns a pointer to the Logger struct. Sever Logger methods were created to output to file and stdOut.
*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	TelnetPort       string `json:"telnetPort"`
	HTTPPort         string `json:"httpPort"`
	MaxClients       int    `json:"maxClients"`
	LogFile          string `json:"logFile"`
	Dialect          string `json:"dialect"`
	ConnectionString string `json:"connectionString"`
}

func Init() {
	if cfgFile == "" {
		GetFlags()
	}

	GetConfig()

	if Logs() == nil {
		Logs()
	}

}

func GetConfig() {
	cfgBody, err := ioutil.ReadFile(cfgFile)
	CheckError(err)
	err = json.Unmarshal(cfgBody, Cfg)
	CheckError(err)
}

func GetFlags() {
	flag.StringVar(&cfgFile, "file", "config/config.json", "config file location")

	flag.Parse()
}

func CheckError(err error) {
	if err != nil {
		Logs().Fatal(fmt.Sprintf("%v. %s", os.Stderr, err.Error()))
	}
}

var cfgFile string

var Cfg *Config = &Config{}
