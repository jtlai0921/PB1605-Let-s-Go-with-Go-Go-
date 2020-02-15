package base

import (
	"net/http"
)

// 資料的接口。
type Data interface {
	Valid() bool // 資料是否有效。
}

// 請求。
type Request struct {
	httpReq *http.Request // HTTP請求的指標值。
	depth   uint32        // 請求的深度。
}

// 建立新的請求。
func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{httpReq: httpReq, depth: depth}
}

// 取得HTTP請求。
func (req *Request) HttpReq() *http.Request {
	return req.httpReq
}

// 取得深度值。
func (req *Request) Depth() uint32 {
	return req.depth
}

// 資料是否有效。
func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
}

// 響應。
type Response struct {
	httpResp *http.Response
	depth    uint32
}

// 建立新的響應。
func NewResponse(httpResp *http.Response, depth uint32) *Response {
	return &Response{httpResp: httpResp, depth: depth}
}

// 取得HTTP響應。
func (resp *Response) HttpResp() *http.Response {
	return resp.httpResp
}

// 取得深度值。
func (resp *Response) Depth() uint32 {
	return resp.depth
}

// 資料是否有效。
func (resp *Response) Valid() bool {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}

// 項目。
type Item map[string]interface{}

// 資料是否有效。
func (item Item) Valid() bool {
	return item != nil
}
