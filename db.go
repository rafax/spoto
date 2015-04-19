package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/mgutz/dat/v1"
	"github.com/mgutz/dat/v1/sqlx-runner"
)

type pinger interface {
	Ping() error
}

var (
	p    pinger
	conn *runner.Connection
)

// Media represents a single image or video on Instagram
type Media struct {
	ID             int       `db:"id"`
	IID            int       `db:"iid"`
	MediaJSON      string    `db:"document"`
	CreatedAt      time.Time `db:"created_at"`
	SubscriptionID int       `db:"subscription_id"`
}

// Subscription represents a named location for which we store Media
type Subscription struct {
	ID     int     `db:"id"`
	Name   string  `db:"name"`
	Lat    float32 `db:"lat"`
	Long   float32 `db:"long"`
	Radius int     `db:"radius"`
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

func insert(media []Media) {
	b := conn.InsertInto("media").Columns("iid", "document", "created_at", "subscription_id").Blacklist("id")
	for _, m := range media {
		b.Record(m)
	}
	_, err := b.Exec()
	if err != nil {
		fmt.Printf("Failed on insert: %v\n", err)
	}
}

func notificationCount() (int, error) {
	var cnt int
	err := conn.SQL("SELECT COUNT(*) FROM media").QueryScalar(&cnt)
	return cnt, err
}
