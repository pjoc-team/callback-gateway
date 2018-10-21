package notify

import (
	"context"
	"flag"
	"gitlab.com/pjoc/base-service/pkg/grpc"
	gc "gitlab.com/pjoc/base-service/pkg/grpc"
	"gitlab.com/pjoc/base-service/pkg/logger"
	"gitlab.com/pjoc/base-service/pkg/service"
	pb "gitlab.com/pjoc/proto/go"
	"time"

	"net/http"
)

const ETCD_DIR_ROOT = "/pjoc/pub/pay"

type NotifyService struct {
	*service.Service
	*grpc.GrpcClientFactory
	*service.GatewayConfig
}

func (svc *NotifyService) Notify(gatewayOrderId string, r *http.Request) (notifyResponse *pb.NotifyResponse, e error) {
	var dbService pb.PayDatabaseServiceClient
	if dbService, e = svc.GetDatabaseClient(); e != nil {
		logger.Log.Errorf("Failed to get db client! error: %v", e.Error())
		return
	}
	timeout, _ := context.WithTimeout(context.TODO(), 10*time.Second)
	orderQuery := &pb.PayOrder{BasePayOrder: &pb.BasePayOrder{GatewayOrderId: gatewayOrderId}}
	var existOrder *pb.PayOrder
	if response, err := dbService.FindPayOrder(timeout, orderQuery); err != nil {
		e = err
		logger.Log.Errorf("Failed to find order! error: %v order: %v", err.Error(), gatewayOrderId)
		return
	} else if response.PayOrders == nil || len(response.PayOrders) == 0 {
		logger.Log.Errorf("Not found order! order: %v", gatewayOrderId)
		return
	} else {
		existOrder = response.PayOrders[0]
	}
	channelId := existOrder.BasePayOrder.ChannelId
	channelAccount := existOrder.BasePayOrder.ChannelAccount

	// send to channel client
	var client pb.PayChannelClient
	if client, e = svc.GetChannelClient(channelId); e != nil {
		logger.Log.Errorf("Failed to get channel client of channelId: %v! error: %v", channelId, e.Error())
		return
	}
	var request *pb.HTTPRequest
	if request, e = BuildChannelHttpRequest(r); e != nil {
		logger.Log.Errorf("Failed to build notify request! error: %v", e.Error())
		return
	}
	notifyRequest := &pb.NotifyRequest{PaymentAccount: channelAccount, Request: request, Type: pb.PayType_PAY, Method: existOrder.BasePayOrder.Method}

	timeoutChannel, _ := context.WithTimeout(context.TODO(), 10*time.Second)
	if notifyResponse, e = client.Notify(timeoutChannel, notifyRequest); e != nil {
		logger.Log.Errorf("Failed to notify channel! error: %v", e.Error())
		return
	}
	return
}

func Init(svc *service.Service) {
	payGatewayService := &NotifyService{}
	payGatewayService.Service = svc
	flag.Parse()

	grpcClientFactory := gc.InitGrpFactory(*svc)
	payGatewayService.GrpcClientFactory = grpcClientFactory

	gatewayConfig := service.InitGatewayConfig(*svc, ETCD_DIR_ROOT)
	payGatewayService.GatewayConfig = gatewayConfig
}
