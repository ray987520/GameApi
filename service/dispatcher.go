package service

import (
	"TestAPI/external/service/mconfig"
	iface "TestAPI/interface"
)

type Dispatcher struct {
	MaxWorkers int                  //最大woker數量
	WorkerPool chan chan iface.IJob //worker jobqueue的channel
	Quit       chan bool
}

// 初始化job queue,worker pool,執行dispatcher
func initDispatcher() {
	MaxWorker = mconfig.GetInt("core.maxWorker")  //最大worker總數,即service服務pool總數
	MaxQueue = mconfig.GetInt("core.maxJobQueue") //最大job queue buffer數
	JobQueue = make(chan iface.IJob, MaxQueue)
	dispatcher := NewDispatcher(MaxWorker)
	dispatcher.Run()
}

// 建立Dispatcher實例,按照最大woker數量宣告對應size的JobChannel
func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan iface.IJob, maxWorkers)
	return &Dispatcher{MaxWorkers: maxWorkers, WorkerPool: pool, Quit: make(chan bool)}
}

// 按MaxWorkers建立worker,另開協程開始分配job
func (d *Dispatcher) Run() {
	for i := 0; i < d.MaxWorkers; i++ {
		//把Dispatcher的JobChannel傳給每一個woker接上
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.Dispatch()
}

// 停止Dispatcher
func (d *Dispatcher) Stop() {
	go func() {
		d.Quit <- true
	}()
}

// Dispatcher分配job
func (d *Dispatcher) Dispatch() {
	for {
		select {
		//取出任一個可用worker的JobChannel,把Job從JobQueue取出後丟進去
		case job := <-JobQueue:
			go func(job iface.IJob) {
				jobChannel := <-d.WorkerPool
				jobChannel <- job
			}(job)
		case <-d.Quit:
			return
		}
	}
}
