package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/kataras/go-sessions"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var err error

type user struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
	Password  string
}

func connecdb() {
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1)/go_db")
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	connecdb()
	router := mux.NewRouter()
	router.PathPrefix("/Login").Handler(http.StripPrefix("/Login", http.FileServer(http.Dir("./Login"))))
	http.ListenAndServe(":7777", router)
}

func login(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) != 0 && checkErr(w, r, err) {
		http.Redirect(w, r, "/", 302)
	}
	if r.Method != "POST" {
		http.ServeFile(w, r, "Login/index.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	users := QueryUser(username)

	//deskripsi dan compare password
	var passwordTes = bcrypt.CompareHashAndPassword([]byte(users.Password), []byte(password))

	if passwordTes == nil {
		//login success
		session := sessions.Start(w, r)
		session.Set("username", users.Username)
		session.Set("name", users.FirstName)
		http.Redirect(w, r, "/", 302)
	} else {
		//login failed
		http.Redirect(w, r, "/login", 302)
	}

}

func home(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	if len(session.GetString("username")) == 0 {
		http.Redirect(w, r, "/login", 301)
	}

	var data = map[string]string{
		"username": session.GetString("username"),
		"message":  "Welcome to the Go !",
	}
	var t, err = template.ParseFiles("views/home.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Execute(w, data)
	return

}

func QueryUser(username string) user {
	var users = user{}
	err = db.QueryRow(`
		SELECT id,
		username,
		first_name,
		last_name,
		password
		FROM users WHERE username=?`, username).
		Scan(
			&users.ID,
			&users.Username,
			&users.FirstName,
			&users.LastName,
			&users.Password,
		)
	return users
}

func checkErr(w http.ResponseWriter, r *http.Request, err error) bool {
	if err != nil {
		fmt.Println(r.Host + r.URL.Path)
		http.Redirect(w, r, r.Host+r.URL.Path, 301)
		return false
	}
	return true
}

func logout(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/", 302)
}
