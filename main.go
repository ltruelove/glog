package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
)

var connectionString string
var dbUser string
var dbPass string
var dbName string
var db *sql.DB

func main() {
	flag.StringVar(&dbUser, "dbUser", "", "Database user")
	flag.StringVar(&dbPass, "dbPass", "", "Database pass")
	flag.StringVar(&dbName, "dbName", "", "Database name")

	flag.Parse()
	connectionString = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True", dbUser, dbPass, dbName)
	var dbErr error
	db, dbErr = sql.Open("mysql", connectionString)
	if dbErr != nil {
		panic(dbErr)
	}
	defer db.Close()

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
