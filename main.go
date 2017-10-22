package main

import (
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
	DefineUserRoutes(router)

	//this has to be last or it will override ports
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))
	http.Handle("/", router)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
