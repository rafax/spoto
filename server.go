package main

import (
	"fmt"
	"net/http"
	"os"

	"runtime"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
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
	var cnt int
	sid := p.ByName("sid")
	sub := getSubscription(sid)
	fetchQueue := make(chan *Media)
	go fetchMedia(sub, fetchQueue)
	for m := range fetchQueue {
		insert(*m)
		cnt++
	}
	fmt.Fprintf(w, "Written %d\n", cnt)
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
