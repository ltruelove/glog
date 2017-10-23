package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	First     string `json:"first"`
	Last      string `json:"last"`
	Pass      string `json:"-"`
	Admin     int    `json:"admin"`
	IsDeleted int    `json:"-"`
}

func DefineUserRoutes(router *mux.Router) {
	router.HandleFunc("/authenticate", GetTokenHandler).Methods("POST")
	router.Handle("/user/{id}", jwtMiddleware.Handler(FetchUser)).Methods("GET")
	router.Handle("/user", jwtMiddleware.Handler(userCreate)).Methods("POST")
	router.Handle("/user", jwtMiddleware.Handler(userUpdate)).Methods("PUT")
	router.Handle("/user/{id}", jwtMiddleware.Handler(Delete)).Methods("DELETE")
	router.Handle("/user/undelete/{id}", jwtMiddleware.Handler(Undelete)).Methods("PUT")
}

func GetUser(id int) (User, error) {
	var user User

	query := fmt.Sprintf(`SELECT id,
    email,
    pass,
    first,
    last,
    admin
    FROM users WHERE id=%d`, id)

	rows, err := db.Query(query)

	for rows.Next() {
		err := rows.Scan(&user.Id,
			&user.Email,
			&user.Pass,
			&user.First,
			&user.Last,
			&user.Admin)

		if err != nil {
			return user, err
		}
	}

	rows.Close()

	if err != nil {
		return user, err
	}

	return user, nil
}

func DeleteUser(id int) error {
	query := fmt.Sprintf(`UPDATE USERS
    SET isDeleted = 1
    WHERE id=%d`, id)

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func UndeleteUser(id int) error {
	query := fmt.Sprintf(`UPDATE USERS
    SET isDeleted = 0
    WHERE id=%d`, id)

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (user *User) Update() error {
	updateStatement, err := db.Prepare(`UPDATE users
    SET 
    first = ?,
    last = ?
    WHERE id = ?`)

	if err != nil {
		return err
	}

	_, err = updateStatement.Exec(user.First,
		user.Last,
		user.Id)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func HashPassword(pass string) string {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), 0)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	return string(hashedPass)
}

func (user *User) Create() (User, error) {
	var newUser User

	insertStatement, err := db.Prepare(`INSERT INTO users
    (email,
    pass,
    first,
    last)
    VALUES (?,?,?,?)`)

	if err != nil {
		return newUser, err
	}

	result, err := insertStatement.Exec(user.Email,
		user.Pass,
		user.First,
		user.Last)

	if err != nil {
		return newUser, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return newUser, err
	}

	newUser.Id = int(id)
	newUser.Email = user.Email
	newUser.First = user.First
	newUser.Last = user.Last

	return newUser, nil
}

func ValidateUser(email string, pass string) (bool, error) {
	var hashedPass string
	query := fmt.Sprintf("SELECT email, pass FROM users WHERE email='%s'", email)
	err := db.QueryRow(query).Scan(&email, &hashedPass)

	if err != nil {
		return false, err
	}

	if bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass)) == nil {
		return true, nil
	}

	return false, nil
}

func AuthUser(writer http.ResponseWriter, request *http.Request) {
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

var FetchUser = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println(err.Error())
	}

	userResult, err := GetUser(id)
	if err != nil {
		fmt.Println(err.Error())
	}

	encodedUser, err := json.Marshal(userResult)
	if err != nil {
		fmt.Println(err.Error())
	}

	writer.Write(encodedUser)
})

var userCreate = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	var user User

	err := decoder.Decode(&user)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	checkUser, err := GetUser(user.Id)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	if checkUser.Id > 0 {
		writer.WriteHeader(400)
		writer.Write([]byte("User record exists."))
		return
	} else {
		user.Pass = HashPassword(user.Pass)
		newUser, err := user.Create()
		if err != nil {
			writer.WriteHeader(500)
			writer.Write([]byte(err.Error()))
			return
		}
		encodedUser, err := json.Marshal(newUser)
		if err != nil {
			writer.WriteHeader(500)
			writer.Write([]byte(err.Error()))
			return
		}

		writer.WriteHeader(201)
		writer.Write(encodedUser)
	}
})

var userUpdate = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	var user User

	err := decoder.Decode(&user)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	checkUser, err := GetUser(user.Id)
	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	if checkUser.Id < 1 {
		writer.WriteHeader(404)
		writer.Write([]byte("User not found"))
		return
	}

	err = user.Update()
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte(err.Error()))
	} else {
		writer.WriteHeader(200)
		writer.Write([]byte("Success"))
	}
})

var Delete = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte(err.Error()))
	}

	err = DeleteUser(id)
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte(err.Error()))
	}

	writer.WriteHeader(200)
	writer.Write([]byte("Delete successful"))
})

var Undelete = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte(err.Error()))
	}

	err = UndeleteUser(id)
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte(err.Error()))
	}

	writer.WriteHeader(200)
	writer.Write([]byte("Undelete successful"))
})

func GetTokenHandler(writer http.ResponseWriter, request *http.Request) {
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
			token := jwt.New(jwt.SigningMethodHS256)
			claims := token.Claims.(jwt.MapClaims)

			claims["email"] = user.Email
			claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

			tokenString, _ := token.SignedString(mySigningKey)

			writer.Write([]byte(tokenString))
		} else {
			writer.WriteHeader(401)
			fmt.Println("Authentication failed")
		}
	}
}
