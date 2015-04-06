package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gocraft/web"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"time"
)

var (
	insertNotification *sql.Stmt
	db                 *sql.DB
)

type Context struct {
}

type Notification struct {
	SubscriptionId string `json:"subscription_id"`
	ObjectId       string `json:"object_id"`
	Object         string `json:"object"`
	ChangedAspect  string `json:"changed_aspect"`
	TimeChanged    int64  `json:"time"`
}

func (c *Context) VerifyInstagram(rw web.ResponseWriter, req *web.Request) {
	req.ParseForm()
	vals, ok := req.Form["hub.challenge"]
	if !ok || len(vals) == 0 {
		fmt.Fprint(rw, "Challenge not found: ", vals)
		return
	}
	challenge := vals[0]
	fmt.Fprint(rw, challenge)
}

func (c *Context) Ping(rw web.ResponseWriter, req *web.Request) {
	err := db.Ping()
	checkErr(err)
	fmt.Fprint(rw, "pong")
}

func saveNotification(n Notification) {
	err := insertNotification.QueryRow(n.SubscriptionId, n.ObjectId, n.ObjectId, n.ChangedAspect, time.Unix(n.TimeChanged, 0)).Scan(&sql.NullInt64{})
	if err != nil {
		fmt.Printf("Failed on insert: %s\n", err)
	}
}

func (c *Context) ReceiveNotifications(rw web.ResponseWriter, req *web.Request) {
	notifications := make([]Notification, 0)
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&notifications)
	for _, n := range notifications {
		go saveNotification(n)
	}
}

func main() {
	dbhost := os.Getenv("SPOTO_DB_HOST")
	if len(dbhost) == 0 {
		dbhost = "localhost"
	}
	cs := fmt.Sprintf("user=spoto password=%s dbname=spoto sslmode=disable host=%s", "otops", dbhost)
	fmt.Println(cs)
	var err error
	db, err = sql.Open("postgres", cs)
	checkErr(err)
	insertNotification, err = db.Prepare("INSERT INTO \"notifications\" (subscription_id, iid, object, changed_aspect, changed_time) VALUES((SELECT id from subscriptions where subscription_id=$1),$2,$3,$4,$5) returning id;")
	checkErr(err)
	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Get("/insta", (*Context).VerifyInstagram).
		Get("/ping", (*Context).Ping).
		Post("/insta", (*Context).ReceiveNotifications)
	http.ListenAndServe("localhost:3000", router)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
