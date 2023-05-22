package es

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

var sonyFlake *sonyflake.Sonyflake

// 取機器ID
func getMachineID() (machineID uint16, err error) {
	//TODO 暫時使用一個假的machineID,後續應有環境變數或其他方式提供機器ID
	machineID = 1688
	return
}

// 初始化,設置sonyFlake基礎值
func init() {
	//需要設置一個固定時間起點讓sonyFlakeID的timestamp區段不重複
	beginTime, err := time.Parse(DbTimeFormat, "2023-01-01 00:00:00.000")
	if err != nil {
		return
	}
	st := sonyflake.Settings{
		StartTime: beginTime,
	}
	st.MachineID = getMachineID
	sonyFlake = sonyflake.NewSonyflake(st)
	return
}

// 產生sonyFlakeID
func Gen() (uuid string, err error) {
	if sonyFlake == nil {
		err = fmt.Errorf("Init SonyFlake err: %#v \n", err)
		return "", err
	}
	id, err := sonyFlake.NextID()
	if err != nil {
		return "", err
	}
	uuid = strconv.FormatUint(id, 16)
	return
}
