package domain

import (
	"TestAPI/entity"
	"TestAPI/enum/errorcode"
	"TestAPI/enum/functionid"
	"TestAPI/enum/innererror"
	es "TestAPI/external/service"
	"TestAPI/external/service/tracer"
	"TestAPI/external/service/zaplog"
	iface "TestAPI/interface"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type DefaultErrorService struct {
	Request entity.DefaultError
}

var validate *validator.Validate

// 初始化驗證器並註冊自訂驗證器
func init() {
	//*TODO 若有需要其他語言錯誤訊息也需要在此添加翻譯元件
	validate = validator.New()
	validate.RegisterValidation("acct", ValidateAccount) //自訂帳號驗證器
}

// 無法找到對應service時由此產生job承載的resquest
func ParseDefaultError(traceMap string, r *http.Request) (request entity.DefaultError) {
	//read request header
	request.Authorization = r.Header.Get(authHeader)
	request.ContentType = r.Header.Get(contentTypeHeader)
	request.TraceID = r.Header.Get(traceHeader)
	request.RequestTime = r.Header.Get(requestTimeHeader)
	request.ErrorCode = r.Header.Get(errorCodeHeader)
	return request
}

// 驗證結構,按struct的validate tag
func IsValid(traceId string, data interface{}) bool {
	err := validate.Struct(data)
	//validate套件驗證到錯誤
	if err != nil {
		zaplog.Errorw(innererror.ValidRequestError, innererror.FunctionNode, functionid.IsValid, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return false
	}

	return err == nil
}

// 自訂驗證帳號驗證器
func ValidateAccount(f1 validator.FieldLevel) bool {
	match, err := regexp.MatchString("[a-zA-Z0-9_-]{3,32}", f1.Field().String())
	if err != nil {
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ValidateAccount, innererror.TraceNode, tracer.DefaultTraceId, innererror.ErrorInfoNode, err)
		return false
	}
	return match
}

func (service *DefaultErrorService) Exec() (data interface{}) {
	//catch panic
	defer tracer.PanicTrace(service.Request.TraceID)
	return nil
}

func (service *DefaultErrorService) GetBaseSelfDefine() entity.BaseSelfDefine {
	return service.Request.BaseSelfDefine
}

// 讀取http request body
func readHttpRequestBody(traceId string, r *http.Request, request iface.IRequest) ([]byte, bool) {
	body, err := ioutil.ReadAll(r.Body)
	//read request body error
	if err != nil {
		zaplog.Errorw(innererror.ServiceError, innererror.FunctionNode, functionid.ReadHttpRequestBody, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		request.SetErrorCode(string(errorcode.BadParameter))
		return nil, false
	}

	return body, true
}

// 解析request body json
func parseJsonBody(traceId string, body []byte, request iface.IRequest) bool {
	isOK := es.JsonUnMarshal(traceId, body, &request)
	//json deserialize error
	if !isOK {
		request.SetErrorCode(string(errorcode.BadParameter))
		return false
	}
	return true
}
