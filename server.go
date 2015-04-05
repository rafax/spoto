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
	insert *sql.Stmt
	db     *sql.DB
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
	chal := vals[0]
	fmt.Fprint(rw, chal)
}

func (c *Context) Ping(rw web.ResponseWriter, req *web.Request) {
	err := db.Ping()
	checkErr(err)
	fmt.Fprint(rw, "pong")
}

func (c *Context) ReceiveNotifications(rw web.ResponseWriter, req *web.Request) {
	notifications := make([]Notification, 0)
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&notifications)
	for _, n := range notifications {
		err := insert.QueryRow(n.SubscriptionId, n.ObjectId, n.ObjectId, n.ChangedAspect, time.Unix(n.TimeChanged, 0)).Scan(&sql.NullInt64{})
		checkErr(err)
	}
	fmt.Fprint(rw, notifications)
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
	insert, err = db.Prepare("INSERT INTO \"notifications\" (subscription_id, iid, object, changed_aspect, changed_time) VALUES((SELECT id from subscriptions where subscription_id=$1),$2,$3,$4,$5) returning id;")
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
