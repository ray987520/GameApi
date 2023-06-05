package domain

import (
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	es "TestAPI/external/service"
	"TestAPI/external/service/zaplog"
	iface "TestAPI/interface"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type DefaultErrorService struct {
	Request  entity.DefaultError
	TraceMap string
}

var validate *validator.Validate

// 初始化驗證器並註冊自訂驗證器
func init() {
	//*TODO 若有需要其他語言錯誤訊息也需要在此添加翻譯元件
	validate = validator.New()
	validate.RegisterValidation("acct", ValidateAccount) //自訂帳號驗證器
}

// 無法找到對應service時由此產生job承載的resquest
func ParseDefaultError(traceMap string, r *http.Request) (request entity.DefaultError, err error) {
	//read request header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	return request, nil
}

// 驗證結構,按struct的validate tag
func IsValid(traceMap string, data interface{}) bool {
	err := validate.Struct(data)
	if err != nil {
		zaplog.Errorw(innererror.ValidRequestError, innererror.FunctionNode, functionid.IsValid, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err)
		return false
	}
	return err == nil
}

// 驗證帳號
func ValidateAccount(f1 validator.FieldLevel) bool {
	match, err := regexp.MatchString("[a-zA-Z0-9_-]{3,32}", f1.Field().String())
	if err != nil {
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ValidateAccount, innererror.ErrorInfoNode, err)
		return false
	}
	return match
}

func (service *DefaultErrorService) Exec() (data interface{}) {
	//catch panic
	defer es.PanicTrace(service.TraceMap)
	return nil
}

func (service *DefaultErrorService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}

// 讀取http request body
func readHttpRequestBody(traceMap string, r *http.Request, request iface.IRequest) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		request.SetErrorCode(string(errorcode.BadParameter))
		return nil, err
	}
	return body, nil
}

// 解析request body json
func parseJsonBody(traceMap string, body []byte, request iface.IRequest) error {
	err := es.JsonUnMarshal(es.AddTraceMap(traceMap, string(esid.JsonUnMarshal)), body, &request)
	if err != nil {
		request.SetErrorCode(string(errorcode.BadParameter))
		return err
	}
	return nil
}
