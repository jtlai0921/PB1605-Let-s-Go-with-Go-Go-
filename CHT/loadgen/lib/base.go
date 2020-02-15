package lib

import (
	"time"
)

// 原生請求的結構。
type RawReq struct {
	Id  int64
	Req []byte
}

// 原生響應的結構。
type RawResp struct {
	Id     int64
	Resp   []byte
	Err    error
	Elapse time.Duration
}

type ResultCode int

// 保留 1 ~ 1000 給酬載承受者使用。
const (
	RESULT_CODE_SUCCESS                         = 0    // 成功。
	RESULT_CODE_WARNING_CALL_TIMEOUT ResultCode = 1001 // 呼叫逾時警示。
	RESULT_CODE_ERROR_CALL           ResultCode = 2001 // 呼叫錯誤。
	RESULT_CODE_ERROR_RESPONSE       ResultCode = 2002 // 響應內容錯誤。
	RESULT_CODE_ERROR_CALEE          ResultCode = 2003 // 被呼叫方（被測軟體）的內定錯誤。
	RESULT_CODE_FATAL_CALL           ResultCode = 3001 // 呼叫過程中發生了致命錯誤！
)

func GetResultCodePlain(code ResultCode) string {
	var codePlain string
	switch code {
	case RESULT_CODE_SUCCESS:
		codePlain = "Success"
	case RESULT_CODE_WARNING_CALL_TIMEOUT:
		codePlain = "Call Timeout Warning"
	case RESULT_CODE_ERROR_CALL:
		codePlain = "Call Error"
	case RESULT_CODE_ERROR_RESPONSE:
		codePlain = "Response Error"
	case RESULT_CODE_ERROR_CALEE:
		codePlain = "Callee Error"
	case RESULT_CODE_FATAL_CALL:
		codePlain = "Call Fatal Error"
	default:
		codePlain = "Unknown result code"
	}
	return codePlain
}

// 呼叫結果的結構。
type CallResult struct {
	Id     int64         // ID。
	Req    RawReq        // 原生請求。
	Resp   RawResp       // 原生響應。
	Code   ResultCode    // 響應程式碼。
	Msg    string        // 結果成因的簡述。
	Elapse time.Duration // 耗時。
}

// 酬載發生器的狀態的型態。
type GenStatus int

const (
	STATUS_ORIGINAL GenStatus = 0
	STATUS_STARTED  GenStatus = 1
	STATUS_STOPPED  GenStatus = 2
)

// 酬載發生器的接口。
type Generator interface {
	// 啟動酬載發生器。
	Start()
	// 停止酬載發生器。
	// 第一個結果值代表已發酬載總數，且僅在第二個結果值為true時有效。
	// 第二個結果值代表是否成功將酬載發生器轉變為已停止狀態。
	Stop() (uint64, bool)
	// 取得狀態。
	Status() GenStatus
}
