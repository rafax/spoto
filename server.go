package main

import (
	"fmt"
	"net/http"
	"os"

	"runtime"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
)

const (
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
		http.Error(w, "Count failed", 500)
	}
	fmt.Fprint(w, "Notifications: ", cnt)
}

func fetch(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	failed := 0
	counter := 0
	sid := p.ByName("sid")
	sub := getSubscription(sid)
	fetchQueue := make(chan *Media)
	stop := make(chan struct{})
	go fetchMedia(sub, fetchQueue, stop)
	for m := range fetchQueue {
		new, err := insert(*m)
		if err != nil {
			fmt.Println("Error encountered %v when inserting", err)
		}
		if !new {
			failed++
			if failed == StopAfterFailedInserts {
				stop <- struct{}{}
				fmt.Printf("Stopping after %d, fetched %d\n", failed, counter)
			}
		} else {
			if failed > 0 {
				fmt.Printf("Found a new media after %d invalid", failed)
			}
			failed = 0
			counter++
		}
	}
	fmt.Fprintf(w, "Fetch completed\n")
}

func initAPI() *negroni.Negroni {
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
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
