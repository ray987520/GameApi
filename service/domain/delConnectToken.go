package domain

import (
	"TestAPI/database"
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/sqlid"
	es "TestAPI/external/service"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type DelConnectTokenService struct {
	Request  entity.DelConnectTokenRequest
	TraceMap string
}

// databinding&validate
func ParseDelConnectTokenRequest(traceMap string, r *http.Request) (request entity.DelConnectTokenRequest, err error) {
	body, err := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &request)
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	if !IsValid(es.AddTraceMap(traceMap, string(functionid.IsValid)), request) {
		request.ErrorCode = string(errorcode.BadParameter)
		return
	}
	return
}

func (service *DelConnectTokenService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	logOutToken(es.AddTraceMap(service.TraceMap, string(functionid.LogOutToken)), &service.Request.BaseSelfDefine, service.Request.Token)
	return
}

func logOutToken(traceMap string, selfDefine *entity.BaseSelfDefine, token string) (isOK bool) {
	now := es.LocalNow(8)
	if isOK = database.DeleteToken(es.AddTraceMap(traceMap, sqlid.DeleteToken.String()), token, now); !isOK {
		es.Error("traceMap:%s ,logOutToken error", traceMap)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
		return
	}
	return
}

func (service *DelConnectTokenService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
