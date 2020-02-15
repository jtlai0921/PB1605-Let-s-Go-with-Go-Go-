package cookie

import (
	"code.google.com/p/go.net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
)

// 建立 http.CookieJar 型態的值。
func NewCookiejar() http.CookieJar {
	options := &cookiejar.Options{PublicSuffixList: &myPublicSuffixList{}}
	cj, _ := cookiejar.New(options)
	return cj
}

// cookiejar.PublicSuffixList 接口的實現型態。
type myPublicSuffixList struct{}

func (psl *myPublicSuffixList) PublicSuffix(domain string) string {
	suffix, _ := publicsuffix.PublicSuffix(domain)
	return suffix
}

func (psl *myPublicSuffixList) String() string {
	return "Web crawler - public suffix list (rev 1.0) power by 'code.google.com/p/go.net/publicsuffix'"
}
