package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/segmentio/ksuid"
	//_ "github.com/jinzhu/gorm/dialects/sqlite"
	//_ "github.com/mattn/go-sqlite3"
)

type Questions struct {
	gorm.Model
	Question string
	Answer   string
	UserId   string
}

var counter int
var httpsServer = false
var eliza = ElizaFromFiles("data/responses.txt", "data/substitutions.txt")

/*
func initializeDatabase() {
	db, err := sql.Open("sqlite3", "eliza.db")
	checkErr(err)
	defer db.Close()

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS QUESTIONS(id INTEGER PRIMARY KEY, user_id VARCHAR(64) , question VARCHAR(128), answer VARCHAR(128), date DATE;")
	checkErr(err)
	_, err1 := stmt.Exec()
	checkErr(err1)
}
*/
func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expiration := time.Now().Add(365 * 24 * time.Hour)
	id := ksuid.New().String()
	cookie := http.Cookie{Name: "cookieId", Value: id, Expires: expiration}
	http.SetCookie(w, &cookie)

	// execute templates
	t := template.Must(template.ParseFiles(path.Join("templates", "index.html")))
	p := map[string]interface{}{
		"Title":  "Talk to Eliza",
		"Viewed": 2}
	t.ExecuteTemplate(w, "index.html", p)
}

// handler for the post funcs
func elizaHandler(w http.ResponseWriter, r *http.Request) {
	/*
		db, err := gorm.Open("sqlite3", "eliza.db")
		if err != nil {
			panic("failed to connect database")
		}
		defer db.Close()
	*/
	userIputMessage := r.FormValue("message")

	lizaAnswer := eliza.RespondTo(userIputMessage)

	//db.Create(&Questions{Question: userIputMessage, Answer: lizaAnswer, UserId: "Prototyping"})

	cookie, _ := r.Cookie("cookieId")

	userCookieId := cookie.Value
	fmt.Println(userCookieId + ": " + userIputMessage)
	fmt.Println(lizaAnswer)
	/*
		// save on db
		db, err := sql.Open("sqlite3", "eliza.db")
		checkErr(err)
		defer db.Close()
		// insert
		stmt, err := db.Prepare("INSERT INTO QUESTIONS(id, user_id, question, answer, date) values(?,?,?, CURRENT_TIMESTAMP)")
		checkErr(err1)
		res, err2 := stmt.Exec(rand.Intn(1000000000), splitSlice[2], splitSlice[0], splitSlice[1])
		checkErr(err2)
	*/
	fmt.Fprintf(w, lizaAnswer) // write data to response
}

func main() {

	//initializeDatabase()
	rand.Seed(time.Now().UnixNano())

	myRouter := mux.NewRouter().StrictSlash(true)
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	myRouter.PathPrefix("/static/").Handler(s)

	myRouter.HandleFunc("/", handler)
	myRouter.HandleFunc("/eliza", elizaHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":80", myRouter))
}

// redirect http to https
func redirect(w http.ResponseWriter, req *http.Request) {

	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
