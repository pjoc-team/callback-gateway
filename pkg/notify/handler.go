package notify

import (
	"github.com/gin-gonic/gin"
	"github.com/pjoc-team/base-service/pkg/logger"
	"net/http"
)

func StartGin(service *NotifyService, listenAddr string) {
	engine := gin.New()
	engine.GET("/notify/:gateway_order_id", handleGatewayOrderIdFunc(service)).
		POST("/notify/:gateway_order_id", handleGatewayOrderIdFunc(service))
	//listenAddr := fmt.Sprintf(":%d", port)
	engine.Run(listenAddr)
}

func handleGatewayOrderIdFunc(service *NotifyService) func(*gin.Context) {
	return func(context *gin.Context) {
		gatewayOrderId := context.Param("gateway_order_id")
		if gatewayOrderId == "" {
			logger.Log.Errorf("No parameter gateway_order_id found! request: %v", context.Params)
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		logger.Log.Infof("Processing notify: %s", gatewayOrderId)
		request := context.Request
		notifyResponse, e := service.Notify(gatewayOrderId, request)
		if e != nil {
			logger.Log.Errorf("Failed to process notify! orderId: %s error: %s", gatewayOrderId, e.Error())
			context.AbortWithStatus(http.StatusBadRequest)
			return
		}

		response := notifyResponse.Response
		headers := response.Header
		context.Status(int(response.Status))
		for name, value := range headers {
			context.Header(name, value)
		}
		if n, e := context.Writer.Write(response.Body); e != nil {
			logger.Log.Errorf("failed to write response! error: %v", e.Error())
		} else {
			logger.Log.Debugf("Success response with size: %d", n)
		}
	}
}
