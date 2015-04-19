package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
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

func stats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var cnt int
	err := countNotifications.QueryRow().Scan(&cnt)
	if err != nil {
		http.Error(w, "Count failed", 500)
	}
	fmt.Fprint(w, "Notifications: ", cnt)
}

func receiveNotifications(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var notifications []notification
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&notifications)
	for _, n := range notifications {
		go insert(n)
	}
}

func initRouter() http.Handler {
	router := httprouter.New()
	router.GET("/insta", verifyInstagram)
	router.POST("/insta", receiveNotifications)
	router.GET("/ping", ping)
	router.GET("/stats", stats)
	return router
}

func main() {
	db := initDb()
	defer db.Close()

	host := getEnvOrDefault("SPOTO_HOST", "localhost")
	port := getEnvOrDefault("SPOTO_PORT", "3000")
	bindTo := fmt.Sprintf("%s:%s", host, port)

	router := initRouter()
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
	n.UseHandler(router)
	n.Run(bindTo)
}

func getEnvOrDefault(key, def string) string {
	env := os.Getenv(key)
	if len(env) == 0 {
		env = def
	}
	return env
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
