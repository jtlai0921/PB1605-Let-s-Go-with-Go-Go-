package scheduler

import (
	"fmt"
	"sync"
	base "webcrawler/base"
)

// 狀態字典。
var statusMap = map[byte]string{
	0: "running",
	1: "closed",
}

// 請求快取的接口型態。
type requestCache interface {
	// 將請求放入請求快取。
	put(req *base.Request) bool
	// 從請求快取取得最早被放入且仍在其中的請求。
	get() *base.Request
	// 獲得請求快取的容量。
	capacity() int
	// 獲得請求快取的實時長度，即：其中的請求的即時數量。
	length() int
	// 關閉請求快取。
	close()
	// 取得請求快取的摘要訊息。
	summary() string
}

// 建立請求快取。
func newRequestCache() requestCache {
	rc := &reqCacheBySlice{
		cache: make([]*base.Request, 0),
	}
	return rc
}

// 請求快取的實現型態。
type reqCacheBySlice struct {
	cache  []*base.Request // 請求的儲存媒體。
	mutex  sync.Mutex      // 互斥鎖。
	status byte            // 快取狀態。0表示正在執行，1表示已關閉。
}

func (rcache *reqCacheBySlice) put(req *base.Request) bool {
	if req == nil {
		return false
	}
	if rcache.status == 1 {
		return false
	}
	rcache.mutex.Lock()
	defer rcache.mutex.Unlock()
	rcache.cache = append(rcache.cache, req)
	return true
}

func (rcache *reqCacheBySlice) get() *base.Request {
	if rcache.length() == 0 {
		return nil
	}
	if rcache.status == 1 {
		return nil
	}
	rcache.mutex.Lock()
	defer rcache.mutex.Unlock()
	req := rcache.cache[0]
	rcache.cache = rcache.cache[1:]
	return req
}

func (rcache *reqCacheBySlice) capacity() int {
	return cap(rcache.cache)
}

func (rcache *reqCacheBySlice) length() int {
	return len(rcache.cache)
}

func (rcache *reqCacheBySlice) close() {
	if rcache.status == 1 {
		return
	}
	rcache.status = 1
}

// 摘要訊息範本。
var summaryTemplate = "status: %s, " + "length: %d, " + "capacity: %d"

func (rcache *reqCacheBySlice) summary() string {
	summary := fmt.Sprintf(summaryTemplate,
		statusMap[rcache.status],
		rcache.length(),
		rcache.capacity())
	return summary
}
