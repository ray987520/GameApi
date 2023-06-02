package entity

//取單一將號httprequest
type GetSequenceNumberRequest struct {
	BaseHttpRequest
	BaseSelfDefine
}

//取單一將號responsedata
type GetSequenceNumberResponse struct {
	SequenceNumber string `json:"sequenceNumber"`
}

//取多將號httprequest
type GetSequenceNumbersRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	Quantity int `json:"quantity" validate:"min=1,max=50"`
}

func (req *GetSequenceNumbersRequest) SetErrorCode(errorCode string) {
	req.ErrorCode = errorCode
}

//取多將號responsedata
type GetSequenceNumbersResponse struct {
	SequenceNumber []string `json:"sequenceNumber"`
}

//取需補注單httprequest
type RoundCheckRequest struct {
	BaseHttpRequest
	BaseSelfDefine
	FromDate string `json:"fromDate" validate:"datetime=2006-01-02T15:04:05.999-07:00"`
	ToDate   string `json:"toDate" validate:"datetime=2006-01-02T15:04:05.999-07:00"`
}

//取需補注單responsedata
type RoundCheckResponse struct {
	RoundCheckList []RoundCheckToken `json:"roundCheckList"`
}

//取需補注單responsedata細項
type RoundCheckToken struct {
	Token              string `json:"connectToken" gorm:"column:connectToken"`
	GameSequenceNumber string `json:"gameSequenceNumber" gorm:"column:gameSequenceNumber"`
}
