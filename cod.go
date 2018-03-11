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
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	//"github.com/segmentio/ksuid"
)

type Questions struct {
	gorm.Model
	Question string
	Answer   string
	UserId   string
}

var counter int
var httpsServer = true
var eliza = ElizaFromFiles("data/responses.txt", "data/substitutions.txt")

func initializeDatabase() {
	db, _ := gorm.Open("sqlite3", "eliza.db")
	defer db.Close()
	db.AutoMigrate(&Questions{})

	if !db.HasTable(&Questions{}) {
		db.CreateTable(&Questions{})
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//expiration := time.Now().Add(365 * 24 * time.Hour)
	//id := ksuid.New().String()
	//cookie := http.Cookie{Name: "cookieId", Value: id, Expires: expiration}
	//http.SetCookie(w, &cookie)

	// execute templates
	t := template.Must(template.ParseFiles(path.Join("templates", "index.html")))
	p := map[string]interface{}{
		"Title":  "Aramis' personal page",
		"Viewed": 0}
	t.ExecuteTemplate(w, "index.html", p)
}

func elizaPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// execute templates
	t := template.Must(template.ParseFiles(path.Join("templates", "eliza.html")))
	p := map[string]interface{}{
		"Title":  "Talk to Eliza",
		"Viewed": 0}
	t.ExecuteTemplate(w, "eliza.html", p)
}

// handler for the post funcs
func elizaHandler(w http.ResponseWriter, r *http.Request) {

	db, err := gorm.Open("sqlite3", "eliza.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	userIputMessage := r.FormValue("message")

	lizaAnswer := eliza.RespondTo(userIputMessage)

	//cookie, _ := r.Cookie("cookieId")

	//userCookieId := cookie.Value
	fmt.Println("someuser: " + ": " + userIputMessage)
	fmt.Println(lizaAnswer)

	question := Questions{Question: userIputMessage, Answer: lizaAnswer, UserId: "user"}
	db.Save(&question)
	fmt.Fprintf(w, lizaAnswer) // write data to response
}

func main() {

	initializeDatabase()
	rand.Seed(time.Now().UnixNano())

	myRouter := mux.NewRouter().StrictSlash(true)
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	myRouter.PathPrefix("/static/").Handler(s)

	myRouter.HandleFunc("/", handler)
	myRouter.HandleFunc("/talkToEliza", elizaPage)
	myRouter.HandleFunc("/eliza", elizaHandler).Methods("POST")

	// redirect every http request to https
	if httpsServer == true {
		go http.ListenAndServe(":80", http.HandlerFunc(redirect))

		if err := http.ListenAndServeTLS(":443", "/etc/letsencrypt/live/www.aramishm.com/cert.pem", "/etc/letsencrypt/live/www.aramishm.com/privkey.pem", myRouter); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	} else {
		log.Fatal(http.ListenAndServe(":8080", myRouter))
	}
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
