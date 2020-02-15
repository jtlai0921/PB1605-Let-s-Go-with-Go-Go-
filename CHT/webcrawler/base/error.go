package base

import (
	"bytes"
	"fmt"
)

// 錯誤型態。
type ErrorType string

// 錯誤型態常數。
const (
	DOWNLOADER_ERROR     ErrorType = "Downloader Error"
	ANALYZER_ERROR       ErrorType = "Analyzer Error"
	ITEM_PROCESSOR_ERROR ErrorType = "Item Processor Error"
)

// 爬蟲錯誤的接口。
type CrawlerError interface {
	Type() ErrorType // 獲得錯誤型態。
	Error() string   // 獲得錯誤提示訊息。
}

// 爬蟲錯誤的實現。
type myCrawlerError struct {
	errType    ErrorType // 錯誤型態。
	errMsg     string    // 錯誤提示訊息。
	fullErrMsg string    // 完整的錯誤提示訊息。
}

// 建立一個新的爬蟲錯誤。
func NewCrawlerError(errType ErrorType, errMsg string) CrawlerError {
	return &myCrawlerError{errType: errType, errMsg: errMsg}
}

// 獲得錯誤型態。
func (ce *myCrawlerError) Type() ErrorType {
	return ce.errType
}

// 獲得錯誤提示訊息。
func (ce *myCrawlerError) Error() string {
	if ce.fullErrMsg == "" {
		ce.genFullErrMsg()
	}
	return ce.fullErrMsg
}

// 產生錯誤提示訊息，並給對應的字段給予值。
func (ce *myCrawlerError) genFullErrMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error: ")
	if ce.errType != "" {
		buffer.WriteString(string(ce.errType))
		buffer.WriteString(": ")
	}
	buffer.WriteString(ce.errMsg)
	ce.fullErrMsg = fmt.Sprintf("%s\n", buffer.String())
	return
}
