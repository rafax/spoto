package main

import (
	"fmt"
	"net/http"
	"os"

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

func initRouter() http.Handler {
	router := httprouter.New()
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
