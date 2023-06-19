package zaplog

import (
	esid "TestAPI/enum/externalserviceid"
	"TestAPI/enum/innererror"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sony/sonyflake"
)

type ElasticService struct {
	Index   string
	TraceId string
}

const (
	elsConnectionError = "elasticsearch connection error:%v"
	elsCreateError     = "elasticsearch create data error:%v"
	elsRedisKey        = "elasticsearch:%s"
	elsDataIndex       = "gameapi"
	sonyFlakeBaseTime  = "2023-01-01 00:00:00.000" //需要設置一個固定時間起點讓sonyFlakeID的timestamp區段不重複
	dbTimeFormat       = "2006-01-02 15:04:05.999"
	defaultTraceId     = "init"
	initFlakeTimeError = "init sonyflake base time error:%v"
	flakeInstanceError = "sonyflake instance error"
)

var (
	esClient                 *elasticsearch.Client
	elsHosts                 []string      //elasticsearch連接的host清單
	elsMaxIdleConnsPerHost   int           //elasticsearch每個host最大閒置連接
	elsResponseHeaderTimeout time.Duration //elasticsearch返回timeout
	elsDialTimeout           time.Duration //elasticsearch連接timeout
	sonyFlake                *sonyflake.Sonyflake
)

func initElastic() {
	elsHosts = zviper.GetStringSlice("elasticsearch.hosts")                              //elasticsearch連接的host清單
	elsMaxIdleConnsPerHost = zviper.GetInt("elasticsearch.maxIdleConnsPerHost")          //elasticsearch每個host最大閒置連接
	elsResponseHeaderTimeout = zviper.GetDuration("elasticsearch.responseHeaderTimeout") //elasticsearch返回timeout
	elsDialTimeout = zviper.GetDuration("elasticsearch.dialTimeout")                     //elasticsearch連接timeout
	cfg := elasticsearch.Config{
		Addresses: elsHosts,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: elsMaxIdleConnsPerHost,
			//ResponseHeaderTimeout: elsResponseHeaderTimeout,
			//DialContext:           (&net.Dialer{Timeout: elsDialTimeout}).DialContext,
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, //TLS安全協議版本1.12比較通用,視情況降版
		},
	}
	var err error
	esClient, err = elasticsearch.NewClient(cfg)
	if err != nil {
		err = fmt.Errorf(elsConnectionError, err)
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.ElasticInit, innererror.TraceNode, defaultTraceId, innererror.ErrorInfoNode, err)
		panic(err) // 連線失敗
	}
	beginTime, err := time.Parse(dbTimeFormat, sonyFlakeBaseTime)
	if err != nil {
		err = fmt.Errorf(initFlakeTimeError, err)
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.ElasticInit, innererror.TraceNode, defaultTraceId, innererror.ErrorInfoNode, err)
		return
	}
	st := sonyflake.Settings{
		StartTime: beginTime,
	}
	st.MachineID = func() (uint16, error) { return 16888, nil }
	sonyFlake = sonyflake.NewSonyflake(st)
}

func NewElasticService() *ElasticService {
	return &ElasticService{
		Index: elsDataIndex,
	}
}

func createIndex(traceId, index, id string, json []byte) (n int, err error) {
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(json),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		err = fmt.Errorf(elsCreateError, err)
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.ElasticCreateIndex, innererror.TraceNode, traceId, innererror.ErrorInfoNode, err)
		return res.StatusCode, err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 || res.StatusCode == 201 || res.StatusCode == 202 {
		return 200, nil
	}
	Infow(innererror.InfoNode, innererror.FunctionNode, esid.ElasticCreateIndex, innererror.TraceNode, traceId, innererror.DataNode, res.Body)
	return res.StatusCode, nil
}

func (els *ElasticService) Write(json []byte) (n int, err error) {
	id := fmt.Sprintf("%s-%s", elsDataIndex, genUuid())
	return createIndex(els.TraceId, els.Index, id, json)
}

// 產生sonyFlakeID
func genUuid() (uuid string) {
	if sonyFlake == nil {
		err := fmt.Errorf(flakeInstanceError)
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.TraceNode, defaultTraceId, innererror.ErrorInfoNode, err)
		return ""
	}
	id, err := sonyFlake.NextID()
	if err != nil {
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.TraceNode, defaultTraceId, innererror.ErrorInfoNode, err)
		return ""
	}
	//轉成16進位數字字串(比較短)
	uuid = strconv.FormatUint(id, 16)
	return uuid
}

func DeleteIndex() {
	req := esapi.IndicesDeleteRequest{
		Index: []string{elsDataIndex},
	}
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	fmt.Println(res)
}
