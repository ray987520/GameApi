package service

import (
	"TestAPI/entity"
	"TestAPI/external/service/mconfig"
	iface "TestAPI/interface"
)

// 定義woker基礎數值,應按實際運作再調整
var (
	MaxWorker = mconfig.GetInt("core.maxWorker")   //最大worker總數,即service服務pool總數
	MaxQueue  = mconfig.GetInt("core.maxJobQueue") //最大job queue buffer數
	JobQueue  chan iface.IJob
)

type Worker struct {
	WorkerPool chan chan iface.IJob
	JobChannel chan iface.IJob
	Quit       chan bool
}

// 建立worker實例
func NewWorker(workPool chan chan iface.IJob) Worker {
	return Worker{
		WorkerPool: workPool,
		JobChannel: make(chan iface.IJob),
		Quit:       make(chan bool),
	}
}

// worker啟動,執行job後回傳response到ResponseMap
func (w Worker) Start() {
	go func() {
		for {
			//把worker的JobChannel接上Dispatcher的WorkerPool
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				data := job.Exec()
				//取出自定義屬性用於response
				selfDefine := job.GetBaseSelfDefine()
				response := GetHttpResponse(selfDefine.ErrorCode, selfDefine.RequestTime, selfDefine.TraceID, data)
				//response回塞到ResponseMap的channel,讓controller接到後輸出
				//sync.Map不能用舊的map[key]方式取值賦值,改用sync.Map.Load取值
				value, isOK := ResponseMap.Load(job.GetBaseSelfDefine().TraceID)
				if !isOK {
					continue
				}
				value.(chan entity.BaseHttpResponse) <- response
				//ResponseMap[job.GetBaseSelfDefine().TraceID] <- response
			case <-w.Quit:
				return
			}
		}
	}()
}

// 停止worker
func (w Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}
