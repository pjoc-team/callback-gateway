package notify

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gitlab.com/pjoc/base-service/pkg/grpc"
	gc "gitlab.com/pjoc/base-service/pkg/grpc"
	"gitlab.com/pjoc/base-service/pkg/logger"
	"gitlab.com/pjoc/base-service/pkg/service"
	pb "gitlab.com/pjoc/proto/go"
	"time"

	"net/http"
)

const ETCD_DIR_ROOT = "/pub/pjoc/pay"

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
		e = fmt.Errorf("not found order: %v", gatewayOrderId)
		return
	} else {
		existOrder = response.PayOrders[0]
	}
	logger.Log.Infof("Processing order notify... order: %v", existOrder)
	// notify
	notifyResponse, e = svc.ProcessChannel(existOrder, r)
	if e != nil {
		return
	}

	settlementClient, e := svc.GetSettlementClient()
	if e != nil {
		logger.Log.Errorf("Failed to get settlement client! error: %v", e.Error())
		return
	} else if settlementClient == nil{
		logger.Log.Errorf("settlementClient is nil!")
		e = errors.New("system error")
		return
	}

	settlementRequest := &pb.SettlementPayOrder{Order: existOrder}
	timeoutSettle, _ := context.WithTimeout(context.TODO(), 10*time.Second)

	settlementResponse, e := settlementClient.ProcessOrderSuccess(timeoutSettle, settlementRequest)
	if e != nil {
		logger.Log.Errorf("Failed to settle order: %v error: %v", existOrder, e.Error())
		return
	} else {
		logger.Log.Infof("Notify order with result: %v", settlementResponse)
	}
	return
}

func (svc *NotifyService) ProcessChannel(existOrder *pb.PayOrder, r *http.Request) (notifyResponse *pb.NotifyResponse, e error) {
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
		logger.Log.Errorf("Failed to notify channel! order: %v error: %v", existOrder, e.Error())
		return
	} else {
		logger.Log.Infof("Notify to channel: %v with result: %v", channelId, notifyResponse)
	}
	return
}

func Init(svc *service.Service) *NotifyService {
	notify := &NotifyService{}
	notify.Service = svc
	flag.Parse()

	gatewayConfig := service.InitGatewayConfig(svc.EtcdPeers, ETCD_DIR_ROOT)
	notify.GatewayConfig = gatewayConfig

	grpcClientFactory := gc.InitGrpFactory(*svc, gatewayConfig)
	notify.GrpcClientFactory = grpcClientFactory

	return notify
}
