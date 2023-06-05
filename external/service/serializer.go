package es

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"TestAPI/external/service/zaplog"
	"encoding/json"
)

// Json序列化
func JsonMarshal(traceMap string, v any) (data []byte, err error) {
	data, err = json.Marshal(v)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.JsonMarshal, innererror.ErrorInfoNode, innererror.JsonMarshalError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "v", v)
		return nil, err
	}
	return data, nil
}

// Json反序列化,v請傳址
func JsonUnMarshal(traceMap string, data []byte, v any) (err error) {
	err = json.Unmarshal(data, v)
	if err != nil {
		zaplog.Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.JsonUnMarshal, innererror.ErrorInfoNode, innererror.JsonUnMarshalError, innererror.TraceNode, traceMap, innererror.ErrorInfoNode, err, "data", data)
	}
	return err
}
