package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"runtime"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
)

const (
	// StopAfterFailedInserts is the number of failed inserts after which we will stop fetching from Instagram.
	// Inserts fail when the image already exists, 50 was chosen to allow for images that are fetched out of order.
	StopAfterFailedInserts = 50
)

func ping(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	err := p.Ping()
	if err != nil {
		http.Error(w, "Ping failed", 500)
	}
	fmt.Fprint(w, "pong")
}

func stats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cnt, err := notificationCount()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	fmt.Fprint(w, "Notifications: ", string(cnt))
}

func fetchMediaForSubscription(sid string, stopAfter int) int {
	sub := getSubscription(sid)
	failed := 0
	counter := 0
	fetchQueue := make(chan *Media)
	stop := make(chan struct{})
	go fetchMedia(sub, fetchQueue, stop)
	for m := range fetchQueue {
		new, err := insert(*m)
		if err != nil {
			fmt.Printf("Error encountered %v when inserting\n", err)
		}
		if !new {
			failed++
			if failed == stopAfter {
				stop <- struct{}{}
			}
		} else {
			failed = 0
			counter++
		}
	}
	return counter
}

func fetch(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sid := p.ByName("sid")
	r.ParseForm()
	stopAfter := StopAfterFailedInserts
	v := r.FormValue("stopAfter")
	if v != "" {
		sa, err := strconv.Atoi(v)
		if err == nil {
			stopAfter = sa
		}
	}
	counter := fetchMediaForSubscription(sid, stopAfter)
	fmt.Fprintf(w, "Fetch completed, fetched %d\n", counter)
}

func initAPI() *negroni.Negroni {
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(), negroni.NewStatic(http.Dir("ui")))
	router := httprouter.New()
	router.GET("/ping", ping)
	router.GET("/stats", stats)
	router.GET("/fetch/:sid", fetch)
	n.UseHandler(router)
	return n
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	db := initDb()
	defer db.Close()

	initClient()

	n := initAPI()

	host := getEnvOrDefault("SPOTO_HOST", "localhost")
	port := getEnvOrDefault("SPOTO_PORT", "3000")
	bindTo := fmt.Sprintf("%s:%s", host, port)

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
