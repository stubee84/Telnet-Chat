package api

/*
	OVERVIEW: This is the web based API which accepts GET and POST HTTP requests to the /chat endpoint. The port used by the server is in config.json.
	GET requests return a message history.
	POST requests send any messages into the chat.
	The acceptedKeys variable is a slice of all form keys for either GET or POST that will be accepted. These reference the chat table columns.
*/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"team-cymru-telnet/config"
	"team-cymru-telnet/db"
	"team-cymru-telnet/models/chat"
	"team-cymru-telnet/server"
	"time"
)

var mux = http.NewServeMux()

func Start() {
	mux.HandleFunc("/chat", messageHandler)

	port := fmt.Sprintf(":%s", config.Cfg.HTTPPort)
	serve := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		if err := serve.ListenAndServe(); err != nil {
			config.Logs().Fatal(err.Error())
		}
	}()

	server.AddUniqueClient("web")

	config.Logs().Info("api server has started")
}

func messageHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/chat" {
		http.Error(res, "404 not found", http.StatusNotFound)
		return
	}

	res.Header().Set("Content-Type", "application/json")

	//calls to ParseForm() would return an empty req.Form variable when sending POST while ParseMultiPartForm(int) successfully initializes the variable
	if err := req.ParseMultipartForm(1024); err != nil {
		config.Logs().Fatal(err.Error())
	}

	if req.Method == "GET" {
		get(res, req, req.Form)
	} else if req.Method == "POST" {
		post(res, req, req.Form)
	}
}

func get(res http.ResponseWriter, req *http.Request, formValues url.Values) {
	acceptedQueryStringParameters := []string{"id", "user", "channel", "message", "recipient", "message_type", "limit"}
	if !validateForm(formValues, acceptedQueryStringParameters) {
		http.Error(res, "400 Bad Request", http.StatusBadRequest)
		return
	}

	chatHistory := []chat.Chat{}
	query := queryBuilder(req.Form)
	db.DB.Conn.Raw(query).Scan(&chatHistory)
	res.WriteHeader(200)
	json.NewEncoder(res).Encode(chatHistory)
}

func post(res http.ResponseWriter, req *http.Request, formValues url.Values) {
	acceptedPostParameters := []string{"channel", "message"}
	if !validateForm(formValues, acceptedPostParameters) {
		http.Error(res, "400 Bad Request", http.StatusBadRequest)
		return
	}

	user := server.User{
		Name:      "web",
		TimeStamp: time.Now().Local().Format(time.Stamp),
	}
	if _, ok := formValues["message"]; !ok {
		http.Error(res, "400 Bad Request", http.StatusBadRequest)
		return
	}
	user.Message = formValues.Get("message")

	if _, ok := formValues["channel"]; ok {
		channelNum, err := strconv.Atoi(formValues.Get("channel"))
		config.CheckError(err)
		user.Channel = channelNum
		if !server.SendToChannel(user) {
			res.WriteHeader(400)
			http.Error(res, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		if !server.SendBroadcast(user) {
			res.WriteHeader(400)
			http.Error(res, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	res.WriteHeader(200)
	json.NewEncoder(res).Encode(`{"Success": "successfully submitted message"`)
}

func validateForm(formValues map[string][]string, parameters []string) bool {
	found := true
	for k := range formValues {
		found = false
		for _, key := range parameters {
			if key == k {
				found = true
			}
		}
		if !found {
			return found
		}
	}
	return found
}

func queryBuilder(formValues url.Values) string {
	query := "SELECT * FROM chat c "
	and := false

	if _, ok := formValues["id"]; ok {
		query += fmt.Sprintf("WHERE id = %s", formValues.Get("id"))
		return query
	}

	if _, ok := formValues["user"]; ok {
		checkAnd(&query, and)
		//user is a unique variable to postgres so in order to query for this column we have to create an alias for the table
		query += fmt.Sprintf("c.user = '%s' ", formValues.Get("user"))
		and = true
	}
	if _, ok := formValues["channel"]; ok {
		checkAnd(&query, and)
		query += fmt.Sprintf("channel = '%s' ", formValues.Get("channel"))
		and = true
	}
	if _, ok := formValues["message_type"]; ok {
		checkAnd(&query, and)
		query += fmt.Sprintf("message_type = '%s' ", formValues.Get("message_type"))
		and = true
	}
	if _, ok := formValues["recipient"]; ok {
		checkAnd(&query, and)
		query += fmt.Sprintf("pm_recipient = '%s' ", formValues.Get("recipient"))
		and = true
	}

	limit := 100
	if _, ok := formValues["limit"]; ok {
		limit = 100
	}
	return query + fmt.Sprintf("limit %d", limit)
}

//And -
func checkAnd(query *string, and bool) {
	if and {
		*query += "and "
	} else {
		*query += "WHERE "
	}
}
