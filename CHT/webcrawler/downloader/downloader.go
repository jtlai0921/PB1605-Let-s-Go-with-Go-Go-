package downloader

import (
	"logging"
	"net/http"
	base "webcrawler/base"
	mdw "webcrawler/middleware"
)

// 日志記錄器。
var logger logging.Logger = base.NewLogger()

// ID產生器。
var downloaderIdGenerator mdw.IdGenerator = mdw.NewIdGenerator()

// 產生並傳回ID。
func genDownloaderId() uint32 {
	return downloaderIdGenerator.GetUint32()
}

// 網頁下載器的接口型態。
type PageDownloader interface {
	Id() uint32                                        // 獲得ID。
	Download(req base.Request) (*base.Response, error) // 根據請求下載網頁並傳回響應。
}

// 建立網頁下載器。
func NewPageDownloader(client *http.Client) PageDownloader {
	id := genDownloaderId()
	if client == nil {
		client = &http.Client{}
	}
	return &myPageDownloader{
		id:         id,
		httpClient: *client,
	}
}

// 網頁下載器的實現型態。
type myPageDownloader struct {
	id         uint32      // ID。
	httpClient http.Client // HTTP用戶端。
}

func (dl *myPageDownloader) Id() uint32 {
	return dl.id
}

func (dl *myPageDownloader) Download(req base.Request) (*base.Response, error) {
	httpReq := req.HttpReq()
	logger.Infof("Do the request (url=%s)... \n", httpReq.URL)
	httpResp, err := dl.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	return base.NewResponse(httpResp, req.Depth()), nil
}
