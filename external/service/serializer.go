package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"encoding/json"
)

// Json序列化
func JsonMarshal(traceMap string, v any) (data []byte) {
	data, err := json.Marshal(v)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.JsonMarshal, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "v", v)
		return nil
	}
	return data
}

// Json反序列化,v請傳址
func JsonUnMarshal(traceMap string, data []byte, v any) (isOK bool) {
	err := json.Unmarshal(data, v)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.JsonUnMarshal, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", string(data))
		return false
	}
	return true
}
