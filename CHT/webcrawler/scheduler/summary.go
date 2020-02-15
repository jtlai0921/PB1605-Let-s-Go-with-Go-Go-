package scheduler

import (
	"bytes"
	"fmt"
	base "webcrawler/base"
)

// 分派器摘要訊息的接口型態。
type SchedSummary interface {
	String() string               // 獲得摘要訊息的一般表示。
	Detail() string               // 取得摘要訊息的詳細表示。
	Same(other SchedSummary) bool // 判斷是否與另一份摘要訊息相同。
}

// 建立分派器摘要訊息。
func NewSchedSummary(sched *myScheduler, prefix string) SchedSummary {
	if sched == nil {
		return nil
	}
	urlCount := len(sched.urlMap)
	var urlDetail string
	if urlCount > 0 {
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		for k, _ := range sched.urlMap {
			buffer.WriteString(prefix)
			buffer.WriteString(prefix)
			buffer.WriteString(k)
			buffer.WriteByte('\n')
		}
		urlDetail = buffer.String()
	} else {
		urlDetail = "\n"
	}
	return &mySchedSummary{
		prefix:              prefix,
		running:             sched.running,
		channelArgs:         sched.channelArgs,
		poolBaseArgs:        sched.poolBaseArgs,
		crawlDepth:          sched.crawlDepth,
		chanmanSummary:      sched.chanman.Summary(),
		reqCacheSummary:     sched.reqCache.summary(),
		dlPoolLen:           sched.dlpool.Used(),
		dlPoolCap:           sched.dlpool.Total(),
		analyzerPoolLen:     sched.analyzerPool.Used(),
		analyzerPoolCap:     sched.analyzerPool.Total(),
		itemPipelineSummary: sched.itemPipeline.Summary(),
		urlCount:            urlCount,
		urlDetail:           urlDetail,
		stopSignSummary:     sched.stopSign.Summary(),
	}
}

// 分派器摘要訊息的實現型態。
type mySchedSummary struct {
	prefix              string            // 前綴。
	running             uint32            // 執行標示。
	channelArgs         base.ChannelArgs  // 通道參數的容器。
	poolBaseArgs        base.PoolBaseArgs // 池基本參數的容器。
	crawlDepth          uint32            // 爬取的最大深度。
	chanmanSummary      string            // 通道管理器的摘要訊息。
	reqCacheSummary     string            // 請求快取的摘要訊息。
	dlPoolLen           uint32            // 網頁下載器池的長度。
	dlPoolCap           uint32            // 網頁下載器池的容量。
	analyzerPoolLen     uint32            // 分析器池的長度。
	analyzerPoolCap     uint32            // 分析器池的容量。
	itemPipelineSummary string            // 項目處理管線的摘要訊息。
	urlCount            int               // 已請求的URL的計數。
	urlDetail           string            // 已請求的URL的詳細訊息。
	stopSignSummary     string            // 停止訊號的摘要訊息。
}

func (ss *mySchedSummary) String() string {
	return ss.getSummary(false)
}

func (ss *mySchedSummary) Detail() string {
	return ss.getSummary(true)
}

// 取得摘要訊息。
func (ss *mySchedSummary) getSummary(detail bool) string {
	prefix := ss.prefix
	template := prefix + "Running: %v \n" +
		prefix + "Channel args: %s \n" +
		prefix + "Pool base args: %s \n" +
		prefix + "Crawl depth: %d \n" +
		prefix + "Channels manager: %s \n" +
		prefix + "Request cache: %s\n" +
		prefix + "Downloader pool: %d/%d\n" +
		prefix + "Analyzer pool: %d/%d\n" +
		prefix + "Item pipeline: %s\n" +
		prefix + "Urls(%d): %s" +
		prefix + "Stop sign: %s\n"
	return fmt.Sprintf(template,
		func() bool {
			return ss.running == 1
		}(),
		ss.channelArgs.String(),
		ss.poolBaseArgs.String(),
		ss.crawlDepth,
		ss.chanmanSummary,
		ss.reqCacheSummary,
		ss.dlPoolLen, ss.dlPoolCap,
		ss.analyzerPoolLen, ss.analyzerPoolCap,
		ss.itemPipelineSummary,
		ss.urlCount,
		func() string {
			if detail {
				return ss.urlDetail
			} else {
				return "<concealed>\n"
			}
		}(),
		ss.stopSignSummary)
}

func (ss *mySchedSummary) Same(other SchedSummary) bool {
	if other == nil {
		return false
	}
	otherSs, ok := interface{}(other).(*mySchedSummary)
	if !ok {
		return false
	}
	if ss.running != otherSs.running ||
		ss.crawlDepth != otherSs.crawlDepth ||
		ss.dlPoolLen != otherSs.dlPoolLen ||
		ss.dlPoolCap != otherSs.dlPoolCap ||
		ss.analyzerPoolLen != otherSs.analyzerPoolLen ||
		ss.analyzerPoolCap != otherSs.analyzerPoolCap ||
		ss.urlCount != otherSs.urlCount ||
		ss.stopSignSummary != otherSs.stopSignSummary ||
		ss.reqCacheSummary != otherSs.reqCacheSummary ||
		ss.poolBaseArgs.String() != otherSs.poolBaseArgs.String() ||
		ss.channelArgs.String() != otherSs.channelArgs.String() ||
		ss.itemPipelineSummary != otherSs.itemPipelineSummary ||
		ss.chanmanSummary != otherSs.chanmanSummary {
		return false
	} else {
		return true
	}
}
