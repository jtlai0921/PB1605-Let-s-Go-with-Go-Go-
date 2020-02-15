package middleware

import (
	"fmt"
	"sync"
)

// 停止訊號的接口型態。
type StopSign interface {
	// 置位停止訊號。相當於發出停止訊號。
	// 若果先前已發出過停止訊號，那麼該方法會傳回false。
	Sign() bool
	// 判斷停止訊號是否已被發出。
	Signed() bool
	// 重設停止訊號。相當於收回停止訊號，並清除所有的停止訊號處理記錄。
	Reset()
	// 處理停止訊號。
	// 參數code應該代表停止訊號處理方的代號。該代號會出現在停止訊號的處理記錄中。
	Deal(code string)
	// 取得某一個停止訊號處理方的處理計數。該處理計數會從對應的停止訊號處理記錄中獲得。
	DealCount(code string) uint32
	// 取得停止訊號被處理的總計數。
	DealTotal() uint32
	// 取得摘要訊息。其中應該包括所有的停止訊號處理記錄。
	Summary() string
}

// 建立停止訊號。
func NewStopSign() StopSign {
	ss := &myStopSign{
		dealCountMap: make(map[string]uint32),
	}
	return ss
}

// 停止訊號的實現型態。
type myStopSign struct {
	rwmutex      sync.RWMutex      // 讀寫鎖。
	signed       bool              // 表示訊號是否已發出的標志位。
	dealCountMap map[string]uint32 // 處理計數的字典。
}

func (ss *myStopSign) Sign() bool {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if ss.signed {
		return false
	}
	ss.signed = true
	return true
}

func (ss *myStopSign) Signed() bool {
	return ss.signed
}

func (ss *myStopSign) Reset() {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	ss.signed = false
	ss.dealCountMap = make(map[string]uint32)
}

func (ss *myStopSign) Deal(code string) {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()
	if !ss.signed {
		return
	}
	if _, ok := ss.dealCountMap[code]; !ok {
		ss.dealCountMap[code] = 1
	} else {
		ss.dealCountMap[code] += 1
	}
}

func (ss *myStopSign) DealCount(code string) uint32 {
	ss.rwmutex.RLock()
	defer ss.rwmutex.Unlock()
	return ss.dealCountMap[code]
}

func (ss *myStopSign) DealTotal() uint32 {
	ss.rwmutex.RLock()
	defer ss.rwmutex.Unlock()
	var total uint32
	for _, v := range ss.dealCountMap {
		total += v
	}
	return total
}

func (ss *myStopSign) Summary() string {
	if ss.signed {
		return fmt.Sprintf("signed: true, dealCount: %v", ss.dealCountMap)
	} else {
		return "signed: false"
	}
}
