package main

import "github.com/carbocation/go-instagram/instagram"

var client *instagram.Client

func initClient() {
	client = instagram.NewClient(nil)
	client.ClientID = "437d47053b9149a59770c6e391ed4a48"
}
