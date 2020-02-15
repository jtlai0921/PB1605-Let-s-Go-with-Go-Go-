package itemproc

import (
	"errors"
	"fmt"
	"sync/atomic"
	base "webcrawler/base"
)

// 項目處理管線的接口型態。
type ItemPipeline interface {
	// 傳送項目。
	Send(item base.Item) []error
	// FailFast方法會傳回一個布爾值。該值表示目前的項目處理管線是否是快速失敗的。
	// 這裡的快速失敗是指：只要對某個項目的處理流程在某一個步驟上出錯，
	// 那麼項目處理管線就會忽略掉後續的所有處理步驟並報告錯誤。
	FailFast() bool
	// 設定是否快速失敗。
	SetFailFast(failFast bool)
	// 獲得已傳送、已接受和已處理的項目的計數值。
	// 更確切地說，作為結果值的切片總會有三個元素值。這三個值會分別代表前述的三個計數。
	Count() []uint64
	// 取得正在被處理的項目的數量。
	ProcessingNumber() uint64
	// 取得摘要訊息。
	Summary() string
}

// 建立項目處理管線。
func NewItemPipeline(itemProcessors []ProcessItem) ItemPipeline {
	if itemProcessors == nil {
		panic(errors.New(fmt.Sprintln("Invalid item processor list!")))
	}
	innerItemProcessors := make([]ProcessItem, 0)
	for i, ip := range itemProcessors {
		if ip == nil {
			panic(errors.New(fmt.Sprintf("Invalid item processor[%d]!\n", i)))
		}
		innerItemProcessors = append(innerItemProcessors, ip)
	}
	return &myItemPipeline{itemProcessors: innerItemProcessors}
}

// 項目處理管線的實現型態。
type myItemPipeline struct {
	itemProcessors   []ProcessItem // 項目處理器的清單。
	failFast         bool          // 表示處理是否需要快速失敗的標志位。
	sent             uint64        // 已被傳送的項目的數量。
	accepted         uint64        // 已被接受的項目的數量。
	processed        uint64        // 已被處理的項目的數量。
	processingNumber uint64        // 正在被處理的項目的數量。
}

func (ip *myItemPipeline) Send(item base.Item) []error {
	atomic.AddUint64(&ip.processingNumber, 1)
	defer atomic.AddUint64(&ip.processingNumber, ^uint64(0))
	atomic.AddUint64(&ip.sent, 1)
	errs := make([]error, 0)
	if item == nil {
		errs = append(errs, errors.New("The item is invalid!"))
		return errs
	}
	atomic.AddUint64(&ip.accepted, 1)
	var currentItem base.Item = item
	for _, itemProcessor := range ip.itemProcessors {
		processedItem, err := itemProcessor(currentItem)
		if err != nil {
			errs = append(errs, err)
			if ip.failFast {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	atomic.AddUint64(&ip.processed, 1)
	return errs
}

func (ip *myItemPipeline) FailFast() bool {
	return ip.failFast
}

func (ip *myItemPipeline) SetFailFast(failFast bool) {
	ip.failFast = failFast
}

func (ip *myItemPipeline) Count() []uint64 {
	counts := make([]uint64, 3)
	counts[0] = atomic.LoadUint64(&ip.sent)
	counts[1] = atomic.LoadUint64(&ip.accepted)
	counts[2] = atomic.LoadUint64(&ip.processed)
	return counts
}

func (ip *myItemPipeline) ProcessingNumber() uint64 {
	return atomic.LoadUint64(&ip.processingNumber)
}

var summaryTemplate = "failFast: %v, processorNumber: %d," +
	" sent: %d, accepted: %d, processed: %d, processingNumber: %d"

func (ip *myItemPipeline) Summary() string {
	counts := ip.Count()
	summary := fmt.Sprintf(summaryTemplate,
		ip.failFast, len(ip.itemProcessors),
		counts[0], counts[1], counts[2], ip.ProcessingNumber())
	return summary
}
