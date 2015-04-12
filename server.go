package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/codegangsta/negroni"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

type pinger interface {
	Ping() error
}

var (
	insertNotification *sql.Stmt
	countNotifications *sql.Stmt
	p                  pinger
)

type notification struct {
	SubscriptionID string `json:"subscription_id"`
	ObjectID       string `json:"object_id"`
	Object         string `json:"object"`
	ChangedAspect  string `json:"changed_aspect"`
	TimeChanged    int64  `json:"time"`
}

func verifyInstagram(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	vals, ok := r.Form["hub.challenge"]
	if !ok || len(vals) == 0 {
		fmt.Fprint(w, "Challenge not found: ", vals)
		return
	}
	challenge := vals[0]
	fmt.Fprint(w, challenge)
}

func ping(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	err := p.Ping()
	if err != nil {
		http.Error(w, "Ping failed", 500)
	}
	fmt.Fprint(w, "pong")
}

func sleep(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	s := p.ByName("sleep")
	sleep, _ := strconv.Atoi(s)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	fmt.Fprint(w, "Slept for ", sleep)
}

func fib(n int) int {
	if n <= 2 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

func fibonacci(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	nstr := p.ByName("n")
	n, _ := strconv.Atoi(nstr)
	fmt.Fprint(w, "", fib(n))
}

func stats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var cnt int
	err := countNotifications.QueryRow().Scan(&cnt)
	if err != nil {
		http.Error(w, "Count failed", 500)
	}
	fmt.Fprint(w, "Notifications: ", cnt)
}

func process(n notification) {
	err := insertNotification.QueryRow(n.SubscriptionID, n.ObjectID, n.Object, n.ChangedAspect, time.Unix(n.TimeChanged, 0)).Scan(&sql.NullInt64{})
	if err != nil {
		fmt.Printf("Failed on insert: %v\n", err)
	}
}

func receiveNotifications(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var notifications []notification
	decoder := json.NewDecoder(r.Body)
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
	countNotifications, err = db.Prepare("SELECT COUNT(*) FROM \"notifications\"")
	checkErr(err)
	p = db
	return db
}

func initRouter(logger, fatal bool) http.Handler {
	router := httprouter.New()
	router.GET("/insta", verifyInstagram)
	router.GET("/ping", ping)
	router.POST("/insta", receiveNotifications)
	router.GET("/stats", stats)
	router.GET("/sleep/:sleep", sleep)
	router.GET("/fib/:n", fibonacci)
	return router
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	db := initDb()
	defer db.Close()

	host := getEnvOrDefault("SPOTO_HOST", "localhost")
	port := getEnvOrDefault("SPOTO_PORT", "3000")
	bindTo := fmt.Sprintf("%s:%s", host, port)

	router := initRouter(false, false)
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
	n.UseHandler(router)
	n.Run(bindTo)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
