package notify

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.com/pjoc/base-service/pkg/logger"
	"gitlab.com/pjoc/notify-gateway/pkg/service"
	"net/http"
)

func StartGin(service *service.NotifyService, port int) {
	engine := gin.New()
	engine.GET("/notify/:gateway_order_id", handleGatewayIdFunc(service)).
		POST("/notify/:gateway_order_id", handleGatewayIdFunc(service))
	listenAddr := fmt.Sprintf(":%d", port)
	engine.Run(listenAddr)
}

func handleGatewayIdFunc(service *service.NotifyService) func(*gin.Context) {
	return func(context *gin.Context) {
		if gatewayOrderId := context.Param("gateway_order_id"); gatewayOrderId == "" {
			logger.Log.Errorf("No parameter gateway_order_id found! request: %v", context.Params)
			context.AbortWithStatus(http.StatusBadRequest)
			return
		} else {
			logger.Log.Infof("Processing notify: %s", gatewayOrderId)
			request := context.Request
			if bytes, e := service.Notify(gatewayOrderId, request); e != nil {
				logger.Log.Errorf("Failed to process notify! orderId: %s error: %s", gatewayOrderId, e.Error())
				context.AbortWithStatus(http.StatusBadRequest)
			} else {
				context.Writer.Write(bytes)
			}
		}
	}
}
