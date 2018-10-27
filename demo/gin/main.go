package main

import (
	"gitlab.com/pjoc/callback-gateway/pkg/notify"
)

func main() {
	notifyService := &notify.NotifyService{}
	notify.StartGin(notifyService, ":8888")
}
