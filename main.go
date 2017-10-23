package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"net/http"
)

var connectionString string
var dbUser string
var dbPass string
var dbName string
var db *sql.DB
var mySigningKey = []byte("secret")

func main() {
	flag.StringVar(&dbUser, "dbUser", "", "Database user")
	flag.StringVar(&dbPass, "dbPass", "", "Database pass")
	flag.StringVar(&dbName, "dbName", "", "Database name")

	flag.Parse()
	connectionString = fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True", dbUser, dbPass, dbName)
	var dbErr error
	db, dbErr = sql.Open("mysql", connectionString)
	if dbErr != nil {
		fmt.Println(dbErr.Error())
	}
	defer db.Close()

	router := mux.NewRouter()
	DefineUserRoutes(router)
	router.Handle("/test", jwtMiddleware.Handler(handleTest)).Methods("GET")
	router.HandleFunc("/get-token", GetTokenHandler).Methods("GET")

	//this has to be last or it will override ports
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))
	http.Handle("/", router)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

var handleTest = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	writer.Write([]byte("Test Successful"))
})

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},

	SigningMethod: jwt.SigningMethodHS256,
})
