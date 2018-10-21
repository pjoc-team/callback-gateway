package service

import (
	"gitlab.com/pjoc/base-service/pkg/service"
	"net/http"
)

type NotifyService struct {
	*service.Service
}

func (svc *NotifyService) Notify(gatewayOrderId string, r *http.Request) (bytes []byte, e error) {
	return
}
