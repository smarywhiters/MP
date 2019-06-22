package main

import (
	"database/sql"
	"fmt"
	_ "html/template"
	"log"
	"net/http"
	"text/template"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kataras/go-sessions"
	_ "github.com/kataras/go-sessions"
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
	fmt.Println("Connected")
}

func routes() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/", home)
	http.HandleFunc("/logout", logout)
}

func main() {
	connecdb()
	routes()
	defer db.Close()

	/*router := mux.NewRouter()
	router.HandleFunc("/login", login)
	router.HandleFunc("/", home)
	/*router.PathPrefix("/Login").Handler(http.StripPrefix("/Login", http.FileServer(http.Dir("./Login"))))*/
	log.Fatal(http.ListenAndServe(":7777", nil))
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
		http.Redirect(w, r, "/", 301)
	}

	var data = map[string]string{
		"username": session.GetString("username"),
		"message":  "Welcome to the Go !",
	}
	var t, err = template.ParseFiles("Login/Home.html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	t.Execute(w, data)
	return

}

func logout(w http.ResponseWriter, r *http.Request) {
	session := sessions.Start(w, r)
	session.Clear()
	sessions.Destroy(w, r)
	http.Redirect(w, r, "/login", 302)
}
