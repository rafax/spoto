package main

import (
	"fmt"
	"log/syslog"
	"net/http"
	"os"
	"strconv"
	"time"

	"runtime"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"

	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
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

func fetch(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	sid := p.ByName("sid")
	subID, _ := strconv.Atoi(sid)
	stopAfter := parseStopAfter(r)
	counter := fetchMediaForSubscription(subID, stopAfter)
	fmt.Fprintf(w, "Fetch completed, fetched %d\n", counter)
}

// Parse stopAfter param to int or return default
func parseStopAfter(r *http.Request) int {
	r.ParseForm()
	stopAfter := StopAfterFailedInserts
	v := r.FormValue("stopAfter")
	if v != "" {
		sa, err := strconv.Atoi(v)
		if err == nil {
			stopAfter = sa
		}
	}
	return stopAfter
}

func fetchAll(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	stopAfter := parseStopAfter(r)
	fetchedForSub := fetchMediaForAllSubscriptions(stopAfter)
	fmt.Fprintf(w, "Fetch completed, fetched %d\n", fetchedForSub)
}

func fetchMediaForSubscription(sid int, stopAfter int) int {
	sub := getSubscription(sid)
	failed := 0
	counter := 0
	fetchQueue := make(chan *Media)
	stop := make(chan struct{})
	go fetchMedia(sub, fetchQueue, stop)
	for m := range fetchQueue {
		new, err := insert(*m)
		if err != nil {
			log.WithField("err", err).WithField("media", m).Error("Error when inserting")
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

func fetchMediaForAllSubscriptions(stopAfter int) map[int]int {
	subs := getSubscriptions()
	log.WithFields(log.Fields{
		"subscriptions": subs,
		"stopAfter":     stopAfter,
	}).Info("Starting fetch all")
	counts := make(map[int]int)
	for _, sub := range subs {
		start := time.Now()
		counts[sub.ID] = fetchMediaForSubscription(sub.ID, stopAfter)
		took := time.Since(start)
		log.WithField("sid", sub.ID).WithField("took", took).Info("Finished fetch for")
	}
	sum := 0
	for _, v := range counts {
		sum += v
	}
	log.WithField("total", sum).Info("Finished fetch all")
	return counts
}

func initAPI() *negroni.Negroni {
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir("ui")))
	router := httprouter.New()
	router.GET("/ping", ping)
	router.GET("/stats", stats)
	router.GET("/fetch/:sid", fetch)
	router.GET("/fetch-all", fetchAll)
	n.UseHandler(router)
	return n
}

func initLog() {
	hook, err := logrus_syslog.NewSyslogHook("udp", "localhost:514", syslog.LOG_INFO, "")
	if err != nil {
		log.Error("Unable to connect to local syslog daemon")
	} else {
		log.AddHook(hook)
	}
	log.SetOutput(os.Stdout)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	db := initDb()
	defer db.Close()

	initClient()

	n := initAPI()

	initLog()

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
