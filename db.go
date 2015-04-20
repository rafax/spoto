package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/mgutz/dat/v1"
	"github.com/mgutz/dat/v1/sqlx-runner"
)

func insert(media Media) error {
	_, err := conn.InsertInto("media").
		Columns("iid", "document", "created_at", "subscription_id").
		Blacklist("id").
		Record(media).
		Exec()
	return err
}

func notificationCount() (int, error) {
	var cnt int
	err := conn.SQL("SELECT COUNT(*) FROM media").QueryScalar(&cnt)
	return cnt, err
}

func getSubscription(sid string) Subscription {
	sub := Subscription{}
	conn.Select("*").From("subscriptions").Where("id = $1", sid).QueryStruct(&sub)
	return sub
}

var (
	p    pinger
	conn *runner.Connection
)

// Media represents a single image or video on Instagram
type Media struct {
	ID             int       `db:"id"`
	IID            string    `db:"iid"`
	MediaJSON      dat.JSON  `db:"document"`
	CreatedAt      time.Time `db:"created_at"`
	SubscriptionID int       `db:"subscription_id"`
}

// Subscription represents a named location for which we store Media
type Subscription struct {
	ID     int     `db:"id"`
	Name   string  `db:"name"`
	Lat    float64 `db:"lat"`
	Lng    float64 `db:"long"`
	Radius float64 `db:"radius"`
}

type pinger interface {
	Ping() error
}

func initDb() *sql.DB {
	dbhost := getEnvOrDefault("SPOTO_DB_HOST", "localhost")
	cs := fmt.Sprintf("user=spoto password=%s dbname=spoto sslmode=disable host=%s", "otops", dbhost)
	db, err := sql.Open("postgres", cs)
	checkErr(err)
	p = db
	dat.EnableInterpolation = true
	conn = runner.NewConnection(db, "postgres")
	return db
}
