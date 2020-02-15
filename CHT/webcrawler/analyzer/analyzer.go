package analyzer

import (
	"errors"
	"fmt"
	"logging"
	"net/url"
	base "webcrawler/base"
	mdw "webcrawler/middleware"
)

// 日志記錄器。
var logger logging.Logger = base.NewLogger()

// ID產生器。
var analyzerIdGenerator mdw.IdGenerator = mdw.NewIdGenerator()

// 產生並傳回ID。
func genAnalyzerId() uint32 {
	return analyzerIdGenerator.GetUint32()
}

// 分析器的接口型態。
type Analyzer interface {
	Id() uint32 // 獲得ID。
	Analyze(
		respParsers []ParseResponse,
		resp base.Response) ([]base.Data, []error) // 根據規則分析響應並傳回請求和項目。
}

// 建立分析器。
func NewAnalyzer() Analyzer {
	return &myAnalyzer{id: genAnalyzerId()}
}

// 分析器的實現型態。
type myAnalyzer struct {
	id uint32 // ID。
}

func (analyzer *myAnalyzer) Id() uint32 {
	return analyzer.id
}

func (analyzer *myAnalyzer) Analyze(
	respParsers []ParseResponse,
	resp base.Response) (dataList []base.Data, errorList []error) {
	if respParsers == nil {
		err := errors.New("The response parser list is invalid!")
		return nil, []error{err}
	}
	httpResp := resp.HttpResp()
	if httpResp == nil {
		err := errors.New("The http response is invalid!")
		return nil, []error{err}
	}
	var reqUrl *url.URL = httpResp.Request.URL
	logger.Infof("Parse the response (reqUrl=%s)... \n", reqUrl)
	respDepth := resp.Depth()

	// 解析HTTP響應。
	dataList = make([]base.Data, 0)
	errorList = make([]error, 0)
	for i, respParser := range respParsers {
		if respParser == nil {
			err := errors.New(fmt.Sprintf("The document parser [%d] is invalid!", i))
			errorList = append(errorList, err)
			continue
		}
		pDataList, pErrorList := respParser(httpResp, respDepth)
		if pDataList != nil {
			for _, pData := range pDataList {
				dataList = appendDataList(dataList, pData, respDepth)
			}
		}
		if pErrorList != nil {
			for _, pError := range pErrorList {
				errorList = appendErrorList(errorList, pError)
			}
		}
	}
	return dataList, errorList
}

// 加入請求值或項目值到清單。
func appendDataList(dataList []base.Data, data base.Data, respDepth uint32) []base.Data {
	if data == nil {
		return dataList
	}
	req, ok := data.(*base.Request)
	if !ok {
		return append(dataList, data)
	}
	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = base.NewRequest(req.HttpReq(), newDepth)
	}
	return append(dataList, req)
}

// 加入錯誤值到清單。
func appendErrorList(errorList []error, err error) []error {
	if err == nil {
		return errorList
	}
	return append(errorList, err)
}
