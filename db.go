package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type pinger interface {
	Ping() error
}

var (
	insertNotification *sql.Stmt
	countNotifications *sql.Stmt
	p                  pinger
)

func initDb() *sql.DB {
	dbhost := getEnvOrDefault("SPOTO_DB_HOST", "localhost")
	cs := fmt.Sprintf("user=spoto password=%s dbname=spoto sslmode=disable host=%s", "otops", dbhost)
	var err error
	db, err := sql.Open("postgres", cs)
	checkErr(err)
	insertNotification, err = db.Prepare("INSERT INTO \"images\" (iid, document, created_time) VALUES($1,$2,$3) returning id;")
	checkErr(err)
	countNotifications, err = db.Prepare("SELECT COUNT(*) FROM \"images\"")
	checkErr(err)
	p = db
	return db
}

func insert(n notification) {
	err := insertNotification.QueryRow(n.SubscriptionID, n.ObjectID, n.Object, n.ChangedAspect, time.Unix(n.TimeChanged, 0)).Scan(&sql.NullInt64{})
	if err != nil {
		fmt.Printf("Failed on insert: %v\n", err)
	}
}
