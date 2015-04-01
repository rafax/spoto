package main

import (
	"appengine"
	"appengine/datastore"
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
	fmt.Fprint(rw, req.Form["hub.challenge"][0])
}

func (c *Context) StoreNotifications(rw web.ResponseWriter, req *web.Request) {
	notifications := make([]Notification, 0)
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&notifications)
	c := appengine.NewContext(r)
	for _, element := range notifications {
		// element is the element from someSlice for where we are
	}
	fmt.Fprint(rw, notifications)
}

func init() {
	router := web.New(Context{}). // Create your router
					Middleware(web.LoggerMiddleware).             // Use some included middleware
					Middleware(web.ShowErrorsMiddleware).         // ...
					Middleware((*Context).SetHelloCount).         // Your own middleware!
					Get("/insta", (*Context).VerifyInstagram).    // Add a route
					Post("/insta", (*Context).StoreNotifications) // Add a route

	http.Handle("/", router)
}
