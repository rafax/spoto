package main

import (
	"errors"
	"fmt"
	"net/http"

	"net/http/httptest"
	"testing"
)

type fakePinger struct {
	ok bool
}

func (p fakePinger) Ping() error {
	if !p.ok {
		return errors.New("HAHA")
	}
	return nil
}

func TestVerify(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://spoto.gajdulewicz.com/insta?hub.challenge=foo", nil)
	rec := httptest.NewRecorder()
	verifyInstagram(rec, r, nil)
	if fmt.Sprintf("%s", rec.Body) != "foo" {
		t.Error("Failed")
	}
}

func TestPongWhenPingSuccessful(t *testing.T) {
	p = fakePinger{ok: true}
	rec := httptest.NewRecorder()
	ping(rec, nil, nil)
	if fmt.Sprintf("%s", rec.Body) != "pong" {
		fmt.Printf("Received: %s\n", rec.Body)
		t.Error("Failed")
	}
}

func TestErrorWhenPingFailed(t *testing.T) {
	p = fakePinger{ok: false}
	rec := httptest.NewRecorder()
	ping(rec, nil, nil)
	if rec.Code != 500 {
		fmt.Println(rec.Body)
		t.Error("Failed")
	}
}
