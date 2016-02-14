// +build integration

package main

import "testing"

func TestDbPing(t *testing.T) {
	db := initDb()
	err := db.Ping()
	if err != nil {
		t.Errorf("Failed ping %s", err)
	}
}
