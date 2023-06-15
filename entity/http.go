package entity

//文件標準HttpResponse結構
type BaseHttpResponse struct {
	Data   interface{}        `json:"data"`
	Status HttpResponseStatus `json:"status"`
}

//文件標準HttpResponse.Status
type HttpResponseStatus struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	DateTime  string `json:"dateTime"`
	TraceCode string `json:"traceCode"`
}

//文件標準Request Header
type BaseHttpRequest struct {
	Authorization string `json:"authorization" validate:"min=1"`
	ContentType   string `json:"contentType"`
}

type BaseSelfDefine struct {
	TraceID     string `json:"traceid" validate:"min=1"`
	RequestTime string `json:"requesttime" validate:"datetime=2006-01-02T15:04:05.999-07:00"`
	ErrorCode   string `json:"errorcode"`
}

// 判斷是否已經有錯誤
func (selfDefine *BaseSelfDefine) HasError() bool {
	//selfDefine.ErrorCode default在middleware會被設為空字串
	return selfDefine.ErrorCode != ""
}
