package main

import (
	"github.com/pjoc-team/callback-gateway/pkg/notify"
)

func main() {
	notifyService := &notify.NotifyService{}
	notify.StartGin(notifyService, ":8888")
}
