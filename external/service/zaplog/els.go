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
	elsDataIndex       = "gameapi%s"
	sonyFlakeBaseTime  = "2023-01-01 00:00:00.000" //需要設置一個固定時間起點讓sonyFlakeID的timestamp區段不重複
	dbTimeFormat       = "2006-01-02 15:04:05.999"
	defaultTraceId     = "init"
	initFlakeTimeError = "init sonyflake base time error:%v"
	flakeInstanceError = "sonyflake instance error"
)

var (
	esClient                 *elasticsearch.Client
	elsHosts                 []string             //elasticsearch連接的host清單
	elsMaxIdleConnsPerHost   int                  //elasticsearch每個host最大閒置連接
	elsResponseHeaderTimeout time.Duration        //elasticsearch返回timeout
	elsDialTimeout           time.Duration        //elasticsearch連接timeout
	sonyFlake                *sonyflake.Sonyflake //用作documentid
)

func initElastic() {
	//從config.json讀取elastic client設定
	elsHosts = zviper.GetStringSlice("elasticsearch.hosts")                              //elasticsearch連接的host清單
	elsMaxIdleConnsPerHost = zviper.GetInt("elasticsearch.maxIdleConnsPerHost")          //elasticsearch每個host最大閒置連接
	elsResponseHeaderTimeout = zviper.GetDuration("elasticsearch.responseHeaderTimeout") //elasticsearch返回timeout,暫時不用,不確定時間單位,好像很容易觸發timeout
	elsDialTimeout = zviper.GetDuration("elasticsearch.dialTimeout")                     //elasticsearch連接timeout,暫時不用,不確定時間單位,好像很容易觸發timeout

	//宣告elastic client config
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
	//連接elasticsearch server
	esClient, err = elasticsearch.NewClient(cfg)
	if err != nil {
		err = fmt.Errorf(elsConnectionError, err)
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.ElasticInit, innererror.TraceNode, defaultTraceId, innererror.DataNode, err)
		panic(err) // 連線失敗
	}

	//初始化SonyFlake,設定基準時間
	beginTime, err := time.Parse(dbTimeFormat, sonyFlakeBaseTime)
	if err != nil {
		err = fmt.Errorf(initFlakeTimeError, err)
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.ElasticInit, innererror.TraceNode, defaultTraceId, innererror.DataNode, err)
		return
	}
	st := sonyflake.Settings{
		StartTime: beginTime,
	}
	//設定SonyFlake MachineID
	st.MachineID = func() (uint16, error) { return 16888, nil }
	sonyFlake = sonyflake.NewSonyflake(st)
}

func NewElasticService() *ElasticService {
	return &ElasticService{
		Index: "gameapi",
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
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.ElasticCreateIndex, innererror.TraceNode, traceId, innererror.DataNode, err)
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
	zone := time.FixedZone("", 8*60*60)
	index := fmt.Sprintf(elsDataIndex, time.Now().In(zone).Format("20060102"))
	id := fmt.Sprintf("%s-%s", index, genUuid())
	return createIndex(els.TraceId, index, id, json)
}

// 產生sonyFlakeID
func genUuid() (uuid string) {
	if sonyFlake == nil {
		err := fmt.Errorf(flakeInstanceError)
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.TraceNode, defaultTraceId, innererror.DataNode, err)
		return ""
	}
	id, err := sonyFlake.NextID()
	if err != nil {
		Errorw(innererror.ExternalServiceError, innererror.FunctionNode, esid.UuidGen, innererror.TraceNode, defaultTraceId, innererror.DataNode, err)
		return ""
	}
	//轉成16進位數字字串(比較短)
	uuid = strconv.FormatUint(id, 16)
	return uuid
}

func DeleteIndex() {
	//zone := time.FixedZone("", 8*60*60)
	req := esapi.IndicesDeleteRequest{
		//Index: []string{fmt.Sprintf(elsDataIndex, time.Now().In(zone).Format("20060102")), "gameapi:20230620"},
		Index: []string{"gameapi20230621"},
	}
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	fmt.Println(res)
}
