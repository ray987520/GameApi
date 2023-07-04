package domain

import (
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type KickTokenService struct {
	Request entity.KickTokenRequest
}

const (
	readRequestBodyError = "can't read request body:%v"
	newHttpRequestError  = "create new http request error:%v"
)

// databinding&validate
func KickTokenRequest(traceId string, r *http.Request) (request entity.KickTokenRequest) {
	body, isOK := readHttpRequestBody(traceId, r, &request)
	//read body error
	if !isOK {
		return request
	}

	isOK = parseJsonBody(traceId, body, &request)
	//json deserialize error
	if !isOK {
		return request
	}

	//read header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(innererror.TraceNode)
	request.RequestTime = r.Header.Get(innererror.RequestTimeNode)
	request.ErrorCode = r.Header.Get(innererror.ErrorCodeNode)

	//validate request
	if !IsValid(traceId, request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return request
	}
	return request
}

func (service *KickTokenService) Exec() (data interface{}) {
	//catch panic
	defer tracer.PanicTrace(service.Request.TraceID)

	if service.Request.HasError() {
		return nil
	}

	service.Request.ErrorCode = string(errorcode.Success)
	return nil
}

func callThirdApi(selfDefine *entity.BaseSelfDefine, httpMethod, url string, header map[string]string, body io.Reader) *http.Response {
	var (
		data []byte
		err  error
	)

	//if not empty body,read and reset body
	if body != nil {
		data, err = io.ReadAll(body)
		//if read body error,return
		if err != nil {
			err = fmt.Errorf(readRequestBodyError, err)
			zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.CallThirdApi, innererror.TraceNode, selfDefine.TraceID, innererror.ErrorInfoNode, err)
			selfDefine.ErrorCode = string(errorcode.UnknowError)
			return nil
		}
		//reset body message
		body = bytes.NewReader(data)
	}

	zaplog.Infow(innererror.ServiceError, innererror.FunctionNode, functionid.CallThirdApi, innererror.TraceNode, selfDefine.TraceID, innererror.DataNode, tracer.MergeMessage("httpMethod", httpMethod, "url", url, "header", header, "body", string(data)))

	//new http request
	req, err := http.NewRequest(httpMethod, url, body)
	if err != nil {
		err = fmt.Errorf(newHttpRequestError, err)
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.CallThirdApi, innererror.TraceNode, selfDefine.TraceID, innererror.ErrorInfoNode, err)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}

	//set http request header
	for k, v := range header {
		req.Header.Set(k, v)
	}

	//call third api
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf(newHttpRequestError, err)
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.CallThirdApi, innererror.TraceNode, selfDefine.TraceID, innererror.ErrorInfoNode, err)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return nil
	}

	return resp
}
