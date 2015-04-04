package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocraft/web"
	"net/http"
)

type Context struct {
}

type Notification struct {
	SubscriptionId string `json:"subscription_id"`
	Object         string `json:"object"`
	ObjectId       string `json:"object_id"`
	ChangedAspect  string `json:"changed_aspect"`
}

func (c *Context) SetHelloCount(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	next(rw, req)
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

func (c *Context) StoreNotifications(rw web.ResponseWriter, req *web.Request) {
	notifications := make([]Notification, 0)
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&notifications)
	fmt.Fprint(rw, notifications)
}

func main() {
	router := web.New(Context{}). // Create your router
					Middleware(web.LoggerMiddleware).             // Use some included middleware
					Middleware(web.ShowErrorsMiddleware).         // ...
					Middleware((*Context).SetHelloCount).         // Your own middleware!
					Get("/insta", (*Context).VerifyInstagram).    // Add a route
					Post("/insta", (*Context).StoreNotifications) // Add a route

	http.ListenAndServe("localhost:3000", router)
}
