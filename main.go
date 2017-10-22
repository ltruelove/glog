package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

var connectionString string
var dbUser string
var dbPass string
var dbName string

func main() {
	flag.StringVar(&dbUser, "dbUser", "", "Database user")
	flag.StringVar(&dbPass, "dbPass", "", "Database pass")
	flag.StringVar(&dbName, "dbName", "", "Database name")

	flag.Parse()
	connectionString = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True", dbUser, dbPass, dbName)

	router := mux.NewRouter()
	router.HandleFunc("/uservalid", TestUser).Methods("POST")

	//this has to be last or it will override ports
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))
	http.Handle("/", router)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func TestUser(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	var user User

	err := decoder.Decode(&user)
	if err != nil {
		panic(err)
	}

	isValid, err := ValidateUser(user.Email, user.Pass)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if isValid {
			fmt.Println("success")
		} else {
			fmt.Println("failed")
		}
	}
}
