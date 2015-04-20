package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/carbocation/go-instagram/instagram"
)

var client *instagram.Client

func initClient() {
	client = instagram.NewClient(nil)
	client.ClientID = "437d47053b9149a59770c6e391ed4a48"
}

func doFetch(p instagram.Parameters, res chan *Media, sid int) error {
	fmt.Printf("Starting fetch for %v\n", p)
	var err error
	im, _, err := client.Media.Search(&p)
	if err != nil {
		return err
	}
	for _, m := range toMedia(im, sid) {
		res <- &m
		fmt.Print(".")
	}
	if len(im) > 0 {
		min, _ := timeRange(im)
		if time.Now().Add(-7*24*time.Hour).Unix() < min {
			p.MaxTimestamp = min
			err = doFetch(p, res, sid)
		}
	}
	return err
}

func fetchMedia(sub Subscription, res chan *Media) error {
	params := instagram.Parameters{Lat: sub.Lat, Lng: sub.Lng, Distance: sub.Radius}
	err := doFetch(params, res, sub.ID)
	close(res)
	return err
}

func toMedia(im []instagram.Media, sid int) []Media {
	media := make([]Media, len(im))
	for i, m := range im {
		mj, _ := json.Marshal(m)
		media[i] = Media{
			IID:            m.ID,
			CreatedAt:      time.Unix(m.CreatedTime, 0),
			MediaJSON:      mj,
			SubscriptionID: sid,
		}
	}
	return media
}

func timeRange(im []instagram.Media) (int64, int64) {
	min, max := im[0].CreatedTime, im[0].CreatedTime
	for _, m := range im {
		if m.CreatedTime < min {
			min = m.CreatedTime
		}
		if m.CreatedTime > max {
			max = m.CreatedTime
		}
	}
	return min, max
}
