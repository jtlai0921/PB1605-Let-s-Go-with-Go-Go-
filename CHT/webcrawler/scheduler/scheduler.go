package scheduler

import (
	"errors"
	"fmt"
	"logging"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	anlz "webcrawler/analyzer"
	base "webcrawler/base"
	dl "webcrawler/downloader"
	ipl "webcrawler/itempipeline"
	mdw "webcrawler/middleware"
)

// 元件的統一代號。
const (
	DOWNLOADER_CODE   = "downloader"
	ANALYZER_CODE     = "analyzer"
	ITEMPIPELINE_CODE = "item_pipeline"
	SCHEDULER_CODE    = "scheduler"
)

// 日志記錄器。
var logger logging.Logger = base.NewLogger()

// 被用來產生HTTP用戶端的函數型態。
type GenHttpClient func() *http.Client

// 分派器的接口型態。
type Scheduler interface {
	// 開啟分派器。
	// 呼叫該方法會使分派器建立和起始化各個元件。在此之後，分派器會啟動爬取流程的執行。
	// 參數channelArgs代表通道參數的容器。
	// 參數poolBaseArgs代表池基本參數的容器。
	// 參數crawlDepth代表了需要被爬取的網頁的最大深度值。深度大於此值的網頁會被忽略。
	// 參數httpClientGenerator代表的是被用來產生HTTP用戶端的函數。
	// 參數respParsers的值應為分析器所需的被用來解析HTTP響應的函數的序列。
	// 參數itemProcessors的值應為需要被置入項目處理管線中的項目處理器的序列。
	// 參數firstHttpReq即代表第一次請求。分派器會以此為起始點開始執行爬取流程。
	Start(channelArgs base.ChannelArgs,
		poolBaseArgs base.PoolBaseArgs,
		crawlDepth uint32,
		httpClientGenerator GenHttpClient,
		respParsers []anlz.ParseResponse,
		itemProcessors []ipl.ProcessItem,
		firstHttpReq *http.Request) (err error)
	// 呼叫該方法會停止分派器的執行。所有處理模組執行的流程都會被中止。
	Stop() bool
	// 判斷分派器是否正在執行。
	Running() bool
	// 獲得錯誤通道。分派器以及各個處理模組執行過程中出現的所有錯誤都會被傳送到該通道。
	// 若該方法的結果值為nil，則說明錯誤通道不可用或分派器已被停止。
	ErrorChan() <-chan error
	// 判斷所有處理模組是否都處於閒置狀態。
	Idle() bool
	// 取得摘要訊息。
	Summary(prefix string) SchedSummary
}

// 建立分派器。
func NewScheduler() Scheduler {
	return &myScheduler{}
}

// 分派器的實現型態。
type myScheduler struct {
	channelArgs   base.ChannelArgs      // 通道參數的容器。
	poolBaseArgs  base.PoolBaseArgs     // 池基本參數的容器。
	crawlDepth    uint32                // 爬取的最大深度。第一次請求的深度為0。
	primaryDomain string                // 主域名。
	chanman       mdw.ChannelManager    // 通道管理器。
	stopSign      mdw.StopSign          // 停止訊號。
	dlpool        dl.PageDownloaderPool // 網頁下載器池。
	analyzerPool  anlz.AnalyzerPool     // 分析器池。
	itemPipeline  ipl.ItemPipeline      // 項目處理管線。
	reqCache      requestCache          // 請求快取。
	urlMap        map[string]bool       // 已請求的URL的字典。
	running       uint32                // 執行標示。0表示未執行，1表示已執行，2表示已停止。
}

func (sched *myScheduler) Start(
	channelArgs base.ChannelArgs,
	poolBaseArgs base.PoolBaseArgs,
	crawlDepth uint32,
	httpClientGenerator GenHttpClient,
	respParsers []anlz.ParseResponse,
	itemProcessors []ipl.ProcessItem,
	firstHttpReq *http.Request) (err error) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Scheduler Error: %s\n", p)
			logger.Fatal(errMsg)
			err = errors.New(errMsg)
		}
	}()
	if atomic.LoadUint32(&sched.running) == 1 {
		return errors.New("The scheduler has been started!\n")
	}
	atomic.StoreUint32(&sched.running, 1)

	if err := channelArgs.Check(); err != nil {
		return err
	}
	sched.channelArgs = channelArgs
	if err := poolBaseArgs.Check(); err != nil {
		return err
	}
	sched.poolBaseArgs = poolBaseArgs
	sched.crawlDepth = crawlDepth

	sched.chanman = generateChannelManager(sched.channelArgs)
	if httpClientGenerator == nil {
		return errors.New("The HTTP client generator list is invalid!")
	}
	dlpool, err :=
		generatePageDownloaderPool(
			sched.poolBaseArgs.PageDownloaderPoolSize(),
			httpClientGenerator)
	if err != nil {
		errMsg :=
			fmt.Sprintf("Occur error when get page downloader pool: %s\n", err)
		return errors.New(errMsg)
	}
	sched.dlpool = dlpool
	analyzerPool, err := generateAnalyzerPool(sched.poolBaseArgs.AnalyzerPoolSize())
	if err != nil {
		errMsg :=
			fmt.Sprintf("Occur error when get analyzer pool: %s\n", err)
		return errors.New(errMsg)
	}
	sched.analyzerPool = analyzerPool

	if itemProcessors == nil {
		return errors.New("The item processor list is invalid!")
	}
	for i, ip := range itemProcessors {
		if ip == nil {
			return errors.New(fmt.Sprintf("The %dth item processor is invalid!", i))
		}
	}
	sched.itemPipeline = generateItemPipeline(itemProcessors)

	if sched.stopSign == nil {
		sched.stopSign = mdw.NewStopSign()
	} else {
		sched.stopSign.Reset()
	}

	sched.reqCache = newRequestCache()
	sched.urlMap = make(map[string]bool)

	sched.startDownloading()
	sched.activateAnalyzers(respParsers)
	sched.openItemPipeline()
	sched.schedule(10 * time.Millisecond)

	if firstHttpReq == nil {
		return errors.New("The first HTTP request is invalid!")
	}
	pd, err := getPrimaryDomain(firstHttpReq.Host)
	if err != nil {
		return err
	}
	sched.primaryDomain = pd

	firstReq := base.NewRequest(firstHttpReq, 0)
	sched.reqCache.put(firstReq)

	return nil
}

func (sched *myScheduler) Stop() bool {
	if atomic.LoadUint32(&sched.running) != 1 {
		return false
	}
	sched.stopSign.Sign()
	sched.chanman.Close()
	sched.reqCache.close()
	atomic.StoreUint32(&sched.running, 2)
	return true
}

func (sched *myScheduler) Running() bool {
	return atomic.LoadUint32(&sched.running) == 1
}

func (sched *myScheduler) ErrorChan() <-chan error {
	if sched.chanman.Status() != mdw.CHANNEL_MANAGER_STATUS_INITIALIZED {
		return nil
	}
	return sched.getErrorChan()
}

func (sched *myScheduler) Idle() bool {
	idleDlPool := sched.dlpool.Used() == 0
	idleAnalyzerPool := sched.analyzerPool.Used() == 0
	idleItemPipeline := sched.itemPipeline.ProcessingNumber() == 0
	if idleDlPool && idleAnalyzerPool && idleItemPipeline {
		return true
	}
	return false
}

func (sched *myScheduler) Summary(prefix string) SchedSummary {
	return NewSchedSummary(sched, prefix)
}

// 開始下載。
func (sched *myScheduler) startDownloading() {
	go func() {
		for {
			req, ok := <-sched.getReqChan()
			if !ok {
				break
			}
			go sched.download(req)
		}
	}()
}

// 下載。
func (sched *myScheduler) download(req base.Request) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Download Error: %s\n", p)
			logger.Fatal(errMsg)
		}
	}()
	downloader, err := sched.dlpool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Downloader pool error: %s", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.dlpool.Return(downloader)
		if err != nil {
			errMsg := fmt.Sprintf("Downloader pool error: %s", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()
	code := generateCode(DOWNLOADER_CODE, downloader.Id())
	respp, err := downloader.Download(req)
	if respp != nil {
		sched.sendResp(*respp, code)
	}
	if err != nil {
		sched.sendError(err, code)
	}
}

// 啟動分析器。
func (sched *myScheduler) activateAnalyzers(respParsers []anlz.ParseResponse) {
	go func() {
		for {
			resp, ok := <-sched.getRespChan()
			if !ok {
				break
			}
			go sched.analyze(respParsers, resp)
		}
	}()
}

// 分析。
func (sched *myScheduler) analyze(respParsers []anlz.ParseResponse, resp base.Response) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Analysis Error: %s\n", p)
			logger.Fatal(errMsg)
		}
	}()
	analyzer, err := sched.analyzerPool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
		sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sched.analyzerPool.Return(analyzer)
		if err != nil {
			errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
			sched.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()
	code := generateCode(ANALYZER_CODE, analyzer.Id())
	dataList, errs := analyzer.Analyze(respParsers, resp)
	if dataList != nil {
		for _, data := range dataList {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *base.Request:
				sched.saveReqToCache(*d, code)
			case *base.Item:
				sched.sendItem(*d, code)
			default:
				errMsg := fmt.Sprintf("Unsupported data type '%T'! (value=%v)\n", d, d)
				sched.sendError(errors.New(errMsg), code)
			}
		}
	}
	if errs != nil {
		for _, err := range errs {
			sched.sendError(err, code)
		}
	}
}

// 開啟項目處理管線。
func (sched *myScheduler) openItemPipeline() {
	go func() {
		sched.itemPipeline.SetFailFast(true)
		code := ITEMPIPELINE_CODE
		for item := range sched.getItemChan() {
			go func(item base.Item) {
				defer func() {
					if p := recover(); p != nil {
						errMsg := fmt.Sprintf("Fatal Item Processing Error: %s\n", p)
						logger.Fatal(errMsg)
					}
				}()
				errs := sched.itemPipeline.Send(item)
				if errs != nil {
					for _, err := range errs {
						sched.sendError(err, code)
					}
				}
			}(item)
		}
	}()
}

// 把請求存放到請求快取。
func (sched *myScheduler) saveReqToCache(req base.Request, code string) bool {
	httpReq := req.HttpReq()
	if httpReq == nil {
		logger.Warnln("Ignore the request! It's HTTP request is invalid!")
		return false
	}
	reqUrl := httpReq.URL
	if reqUrl == nil {
		logger.Warnln("Ignore the request! It's url is is invalid!")
		return false
	}
	if strings.ToLower(reqUrl.Scheme) != "http" {
		logger.Warnf("Ignore the request! It's url scheme '%s', but should be 'http'!\n", reqUrl.Scheme)
		return false
	}
	if _, ok := sched.urlMap[reqUrl.String()]; ok {
		logger.Warnf("Ignore the request! It's url is repeated. (requestUrl=%s)\n", reqUrl)
		return false
	}
	if pd, _ := getPrimaryDomain(httpReq.Host); pd != sched.primaryDomain {
		logger.Warnf("Ignore the request! It's host '%s' not in primary domain '%s'. (requestUrl=%s)\n",
			httpReq.Host, sched.primaryDomain, reqUrl)
		return false
	}
	if req.Depth() > sched.crawlDepth {
		logger.Warnf("Ignore the request! It's depth %d greater than %d. (requestUrl=%s)\n",
			req.Depth(), sched.crawlDepth, reqUrl)
		return false
	}
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.reqCache.put(&req)
	sched.urlMap[reqUrl.String()] = true
	return true
}

// 傳送響應。
func (sched *myScheduler) sendResp(resp base.Response, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getRespChan() <- resp
	return true
}

// 傳送項目。
func (sched *myScheduler) sendItem(item base.Item, code string) bool {
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	sched.getItemChan() <- item
	return true
}

// 傳送錯誤。
func (sched *myScheduler) sendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errorType base.ErrorType
	switch codePrefix {
	case DOWNLOADER_CODE:
		errorType = base.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = base.ANALYZER_ERROR
	case ITEMPIPELINE_CODE:
		errorType = base.ITEM_PROCESSOR_ERROR
	}
	cError := base.NewCrawlerError(errorType, err.Error())
	if sched.stopSign.Signed() {
		sched.stopSign.Deal(code)
		return false
	}
	go func() {
		sched.getErrorChan() <- cError
	}()
	return true
}

// 分派。適當的搬運請求快取中的請求到請求通道。
func (sched *myScheduler) schedule(interval time.Duration) {
	go func() {
		for {
			if sched.stopSign.Signed() {
				sched.stopSign.Deal(SCHEDULER_CODE)
				return
			}
			remainder := cap(sched.getReqChan()) - len(sched.getReqChan())
			var temp *base.Request
			for remainder > 0 {
				temp = sched.reqCache.get()
				if temp == nil {
					break
				}
				if sched.stopSign.Signed() {
					sched.stopSign.Deal(SCHEDULER_CODE)
					return
				}
				sched.getReqChan() <- *temp
				remainder--
			}
			time.Sleep(interval)
		}
	}()
}

// 取得通道管理器持有的請求通道。
func (sched *myScheduler) getReqChan() chan base.Request {
	reqChan, err := sched.chanman.ReqChan()
	if err != nil {
		panic(err)
	}
	return reqChan
}

// 取得通道管理器持有的響應通道。
func (sched *myScheduler) getRespChan() chan base.Response {
	respChan, err := sched.chanman.RespChan()
	if err != nil {
		panic(err)
	}
	return respChan
}

// 取得通道管理器持有的項目通道。
func (sched *myScheduler) getItemChan() chan base.Item {
	itemChan, err := sched.chanman.ItemChan()
	if err != nil {
		panic(err)
	}
	return itemChan
}

// 取得通道管理器持有的錯誤通道。
func (sched *myScheduler) getErrorChan() chan error {
	errorChan, err := sched.chanman.ErrorChan()
	if err != nil {
		panic(err)
	}
	return errorChan
}
