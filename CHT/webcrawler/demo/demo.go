package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"logging"
	"net/http"
	"net/url"
	"strings"
	"time"
	"webcrawler/analyzer"
	base "webcrawler/base"
	pipeline "webcrawler/itempipeline"
	sched "webcrawler/scheduler"
	"webcrawler/tool"
)

// 日志記錄器。
var logger logging.Logger = logging.NewSimpleLogger()

// 項目處理器。
func processItem(item base.Item) (result base.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}
	// 產生結果
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil
}

// 響應解析函數。只解析“A”標簽。
func parseForATag(httpResp *http.Response, respDepth uint32) ([]base.Data, []error) {
	// TODO 支援更多的HTTP響應狀態
	if httpResp.StatusCode != 200 {
		err := errors.New(
			fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", httpResp))
		return nil, []error{err}
	}
	var reqUrl *url.URL = httpResp.Request.URL
	var httpRespBody io.ReadCloser = httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()
	dataList := make([]base.Data, 0)
	errs := make([]error, 0)
	// 開始解析
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}
	// 查詢“A”標簽並分析連結位址
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		// 前期過濾
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}
		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)
		// 暫不支援對Javascript程式碼的解析。
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aUrl, err := url.Parse(href)
			if err != nil {
				errs = append(errs, err)
				return
			}
			if !aUrl.IsAbs() {
				aUrl = reqUrl.ResolveReference(aUrl)
			}
			httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := base.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		}
		text := strings.TrimSpace(sel.Text())
		if text != "" {
			imap := make(map[string]interface{})
			imap["parent_url"] = reqUrl
			imap["a.text"] = text
			imap["a.index"] = index
			item := base.Item(imap)
			dataList = append(dataList, &item)
		}
	})
	return dataList, errs
}

// 獲得響應解析函數的序列。
func getResponseParsers() []analyzer.ParseResponse {
	parsers := []analyzer.ParseResponse{
		parseForATag,
	}
	return parsers
}

// 獲得項目處理器的序列。
func getItemProcessors() []pipeline.ProcessItem {
	itemProcessors := []pipeline.ProcessItem{
		processItem,
	}
	return itemProcessors
}

// 產生HTTP用戶端。
func genHttpClient() *http.Client {
	return &http.Client{}
}

func record(level byte, content string) {
	if content == "" {
		return
	}
	switch level {
	case 0:
		logger.Infoln(content)
	case 1:
		logger.Warnln(content)
	case 2:
		logger.Infoln(content)
	}
}

func main() {
	// 建立分派器
	scheduler := sched.NewScheduler()

	// 準備監控參數
	intervalNs := 10 * time.Millisecond
	maxIdleCount := uint(1000)
	// 開始監控
	checkCountChan := tool.Monitoring(
		scheduler,
		intervalNs,
		maxIdleCount,
		true,
		false,
		record)

	// 準備啟動參數
	channelArgs := base.NewChannelArgs(10, 10, 10, 10)
	poolBaseArgs := base.NewPoolBaseArgs(3, 3)
	crawlDepth := uint32(1)
	httpClientGenerator := genHttpClient
	respParsers := getResponseParsers()
	itemProcessors := getItemProcessors()
	startUrl := "http://www.sogou.com"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		logger.Errorln(err)
		return
	}
	// 開啟分派器
	scheduler.Start(
		channelArgs,
		poolBaseArgs,
		crawlDepth,
		httpClientGenerator,
		respParsers,
		itemProcessors,
		firstHttpReq)

	// 等待監控結束
	<-checkCountChan
}
