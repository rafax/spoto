package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/mgutz/dat/v1"
	"github.com/mgutz/dat/v1/sqlx-runner"
)

func insert(media Media) (bool, error) {
	var exists int64
	conn.Select("1").From("media").Where("iid = $1", media.IID).QueryScalar(&exists)
	if exists > 0 {
		return false, nil
	}
	_, err := conn.Insect("media").
		Columns("iid", "document", "created_at", "subscription_id").
		Blacklist("id").
		Record(media).
		Where("iid = $1", media.IID).
		Exec()
	return true, err
}

func notificationCount() ([]byte, error) {
	json, err := conn.SQL(`
	SELECT subscription_id,
	       s.name,
	       COUNT(iid)
	FROM media m
	JOIN subscriptions s ON m.subscription_id = s.id
	GROUP BY subscription_id,
	         s.name
	ORDER BY subscription_id
		`).QueryJSON()
	return json, err
}

func getSubscription(sid int) Subscription {
	sub := Subscription{}
	conn.Select("*").From("subscriptions").Where("id = $1", sid).QueryStruct(&sub)
	return sub
}

func getSubscriptions() []Subscription {
	var subs []Subscription
	conn.Select("*").From("subscriptions").QueryStructs(&subs)
	return subs
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
