package lib

import (
	"time"
)

// 呼叫器的接口。
type Caller interface {
	// 建構請求。
	BuildReq() RawReq
	// 呼叫。
	Call(req []byte, timeoutNs time.Duration) ([]byte, error)
	// 檢查響應。
	CheckResp(rawReq RawReq, rawResp RawResp) *CallResult
}
