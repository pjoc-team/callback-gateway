package main

import (
	"gitlab.com/pjoc/notify-gateway/pkg/notify"
	"gitlab.com/pjoc/notify-gateway/pkg/service"
)

func main() {
	notifyService := &service.NotifyService{}
	notify.StartGin(notifyService, 8888)
}
