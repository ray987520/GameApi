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

type UpdateTokenLocationService struct {
	Request  entity.UpdateTokenLocationRequest
	TraceMap string
}

// databinding&validate
func ParseUpdateTokenLocationRequest(traceMap string, r *http.Request) (request entity.UpdateTokenLocationRequest, err error) {
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

func (service *UpdateTokenLocationService) Exec() (data interface{}) {
	if service.Request.HasError() {
		return
	}
	if isConnectTokenError(es.AddTraceMap(service.TraceMap, string(functionid.IsConnectTokenError)), &service.Request.BaseSelfDefine, service.Request.Token) {
		return
	}
	isOK := updateTokenLocation(es.AddTraceMap(service.TraceMap, string(functionid.UpdateTokenLocation)), &service.Request.BaseSelfDefine, service.Request.Token, service.Request.Location)
	if !isOK {
		return
	}
	return
}

// 更新connectToken location
func updateTokenLocation(traceMap string, selfDefine *entity.BaseSelfDefine, token string, location int) (isOK bool) {
	isOK = database.UpdateTokenLocation(es.AddTraceMap(traceMap, sqlid.UpdateTokenLocation.String()), token, location)
	if !isOK {
		es.Error("traceMap:%s ,updateTokenLocation error", traceMap)
		selfDefine.ErrorCode = string(errorcode.UnknowError)
	}
	return
}

// 檢查aes token活耀
func isConnectTokenAlive(traceMap string, token string) bool {
	alive := database.GetTokenAlive(es.AddTraceMap(traceMap, sqlid.GetTokenAlive.String()), token)
	if !alive {
		es.Error("tranceMap:%s ,token is dead,token:%s", traceMap, token)
	}
	return alive
}

func (service *UpdateTokenLocationService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
