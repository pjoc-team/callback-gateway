package notify

import (
	"bytes"
	"gitlab.com/pjoc/base-service/pkg/logger"
	pb "gitlab.com/pjoc/proto/go"
	"net/http"
)

func BuildChannelNotifyRequest(r *http.Request){

}


func BuildChannelHttpRequest(r *http.Request) (request *pb.HTTPRequest, err error) {
	var body []byte
	if body, err = GetBody(r); err != nil {
		return nil, err
	}
	request = &pb.HTTPRequest{}
	request.Body = body
	switch r.Method {
	case http.MethodGet:
		request.Method = pb.HTTPRequest_GET
	case http.MethodPost:
		request.Method = pb.HTTPRequest_POST
	default:
		logger.Log.Warnf("unknown http method: %v", r.Method)
		request.Method = pb.HTTPRequest_POST
	}
	request.Header = GetHeader(r)
	return
}

func GetHeader(r http.Request) map[string]string {
	header := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			header[k] = v[0]
		}
	}
	return header
}

func GetBody(r *http.Request) ([]byte, error) {
	body := r.Body
	defer body.Close()
	buffer := bytes.Buffer{}
	if n, err := buffer.ReadFrom(body); err != nil {
		logger.Log.Errorf("Failed when read body! error: %v", err.Error())
		return nil, err
	} else {
		logger.Log.Debugf("Read byte size: %d", n)
		return buffer.Bytes(), nil
	}

}
