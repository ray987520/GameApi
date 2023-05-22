package domain

import (
	"TestAPI/entity"
	es "TestAPI/external/service"
	"net/http"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	//初始化驗證器並註冊自訂驗證器
	//TODO 若有需要其他語言錯誤訊息也需要在此添加翻譯元件
	validate = validator.New()
	validate.RegisterValidation("acct", ValidateAccount)
}

type DefaultErrorService struct {
	Request  entity.DefaultError
	TraceMap string
}

// 無法找到對應service時由此產生job承載的resquest
func ParseDefaultError(traceMap string, r *http.Request) (request entity.DefaultError, err error) {
	request.Authorization = r.Header.Get("Authorization")
	request.ContentType = r.Header.Get("Content-Type")
	request.TraceID = r.Header.Get("traceid")
	request.RequestTime = r.Header.Get("requesttime")
	request.ErrorCode = r.Header.Get("errorcode")
	return
}

// validate struct
func IsValid(traceMap string, data interface{}) bool {
	err := validate.Struct(data)
	if err != nil {
		es.Error("traceMap:%s ,err:%v", traceMap, err)
		return false
	}
	return err == nil
}

// 驗證帳號
func ValidateAccount(f1 validator.FieldLevel) bool {
	match, err := regexp.MatchString("[a-zA-Z0-9_-]{3,32}", f1.Field().String())
	if err != nil {
		es.Error("ValidateAccount err:%v", err)
		return false
	}
	return match
}

func (service *DefaultErrorService) Exec() (data interface{}) {
	data = ""
	return
}

func (service *DefaultErrorService) GetBaseSelfDefine() (selfDefine entity.BaseSelfDefine) {
	return service.Request.BaseSelfDefine
}
