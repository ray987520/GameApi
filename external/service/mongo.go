package es

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 測試mongodb物件
type TimePorint struct {
	StartTime int64 `bson:"startTime"` //开始时间
	EndTime   int64 `bson:"endTime"`   //结束时间
}

// 測試mongodb物件
type LogRecord struct {
	JobName string     `bson:"jobName"` //任务名
	Command string     `bson:"command"` //shell命令
	Err     string     `bson:"err"`     //脚本错误
	Content string     `bson:"content"` //脚本输出
	Tp      TimePorint //执行时间
}

var mgoCli *mongo.Client

// 初始化mongodb
func init() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// 連接MongoDB
	mgoCli, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// 確認連接狀態
	err = mgoCli.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// 測試批次新增50000筆資料
func TestAddMongo() {
	var (
		id primitive.ObjectID
	)
	collection := mgoCli.Database("db").Collection("my_collection")
	//每次5筆,10000次
	for i := 1; i <= 10000; i++ {
		result, err := collection.InsertMany(context.TODO(), []interface{}{
			LogRecord{
				JobName: fmt.Sprintf("job%d", i+0),
				Command: fmt.Sprintf("cmd %d", i+0),
				Err:     "",
				Content: "",
				Tp: TimePorint{
					StartTime: UtcNow().Unix(),
					EndTime:   UtcNow().Unix() + 10,
				},
			},
			LogRecord{
				JobName: fmt.Sprintf("job%d", i+1),
				Command: fmt.Sprintf("cmd %d", i+1),
				Err:     "",
				Content: "",
				Tp: TimePorint{
					StartTime: UtcNow().Unix(),
					EndTime:   UtcNow().Unix() + 10,
				},
			},
			LogRecord{
				JobName: fmt.Sprintf("job%d", i+2),
				Command: fmt.Sprintf("cmd %d", i+2),
				Err:     "",
				Content: "",
				Tp: TimePorint{
					StartTime: UtcNow().Unix(),
					EndTime:   UtcNow().Unix() + 10,
				},
			},
			LogRecord{
				JobName: fmt.Sprintf("job%d", i+3),
				Command: fmt.Sprintf("cmd %d", i+3),
				Err:     "",
				Content: "",
				Tp: TimePorint{
					StartTime: UtcNow().Unix(),
					EndTime:   UtcNow().Unix() + 10,
				},
			},
			LogRecord{
				JobName: fmt.Sprintf("job%d", i+4),
				Command: fmt.Sprintf("cmd %d", i+4),
				Err:     "",
				Content: "",
				Tp: TimePorint{
					StartTime: UtcNow().Unix(),
					EndTime:   UtcNow().Unix() + 10,
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		if result == nil {
			log.Fatal("result nil")
		}
		for _, v := range result.InsertedIDs {
			id = v.(primitive.ObjectID)
			fmt.Println("自增ID", id.Hex())
		}
	}
}

// 測試撈mongodb 10000筆
func TestGetMongo() {
	var (
		err        error
		collection *mongo.Collection
		cursor     *mongo.Cursor
	)
	stime := UtcNow().Format(ApiTimeFormat)
	collection = mgoCli.Database("db").Collection("my_collection")

	//不跳過抓10000筆,無條件
	if cursor, err = collection.Find(context.TODO(), bson.D{}, options.Find().SetSkip(0), options.Find().SetLimit(10000)); err != nil {
		fmt.Println(err)
		return
	}
	//最後要關閉cursor
	defer func() {
		if err = cursor.Close(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()
	/*
		//自行遍歷
		for cursor.Next(context.TODO()) {
			var lr LogRecord
			//反序列化Bson
			if cursor.Decode(&lr) != nil {
				fmt.Print(err)
				return
			}
			fmt.Println(lr)
		}
	*/
	//內建遍歷：
	var results []LogRecord
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	etime := UtcNow().Format(ApiTimeFormat)
	fmt.Println(fmt.Sprintf("start:%s ,end:%s, len:%d", stime, etime, len(results)))
	/*
		for _, result := range results {
			fmt.Println(result)
		}
		fmt.Println(UtcNow().Format(ApiTimeFormat))
	*/
}
