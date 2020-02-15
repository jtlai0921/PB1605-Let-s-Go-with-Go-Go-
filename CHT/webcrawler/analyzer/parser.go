package analyzer

import (
	"net/http"
	base "webcrawler/base"
)

// 被用於解析HTTP響應的函數型態。
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]base.Data, []error)
