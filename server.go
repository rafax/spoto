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

type Pinger interface {
	Ping() error
}

var (
	insertNotification *sql.Stmt
	pinger             Pinger
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
	err := pinger.Ping()
	if err != nil {
		http.Error(rw, "Ping failed", 500)
	}
	fmt.Fprint(rw, "pong")
}

func process(n Notification) {
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
		go process(n)
	}
}

func getEnvOrDefault(key, def string) string {
	env := os.Getenv(key)
	if len(env) == 0 {
		env = def
	}
	return env
}

func initDb() *sql.DB {
	dbhost := getEnvOrDefault("SPOTO_DB_HOST", "localhost")
	cs := fmt.Sprintf("user=spoto password=%s dbname=spoto sslmode=disable host=%s", "otops", dbhost)
	var err error
	db, err := sql.Open("postgres", cs)
	checkErr(err)
	insertNotification, err = db.Prepare("INSERT INTO \"notifications\" (subscription_id, iid, object, changed_aspect, changed_time) VALUES((SELECT id from subscriptions where subscription_id=$1),$2,$3,$4,$5) returning id;")
	checkErr(err)
	pinger = db
	return db
}

func initRouter(logger, fatal bool) http.Handler {
	ctx := web.New(Context{})
	if logger {
		ctx = ctx.Middleware(web.LoggerMiddleware)
	}
	if fatal {
		ctx = ctx.Middleware(web.ShowErrorsMiddleware)
	}
	return ctx.Get("/insta", (*Context).VerifyInstagram).
		Get("/ping", (*Context).Ping).
		Post("/insta", (*Context).ReceiveNotifications)
}

func main() {
	db := initDb()
	defer db.Close()

	host := getEnvOrDefault("SPOTO_HOST", "localhost")
	port := getEnvOrDefault("SPOTO_PORT", "3000")
	bindTo := fmt.Sprintf("%s:%s", host, port)

	router := initRouter(true, true)
	http.ListenAndServe(bindTo, router)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
