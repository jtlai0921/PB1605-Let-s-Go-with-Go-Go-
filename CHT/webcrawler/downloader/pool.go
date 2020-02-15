package downloader

import (
	"errors"
	"fmt"
	"reflect"
	mdw "webcrawler/middleware"
)

// 產生網頁下載器的函數型態。
type GenPageDownloader func() PageDownloader

// 網頁下載器池的接口型態。
type PageDownloaderPool interface {
	Take() (PageDownloader, error)  // 從池中取出一個網頁下載器。
	Return(dl PageDownloader) error // 把一個網頁下載器歸還給池。
	Total() uint32                  // 獲得池的總容量。
	Used() uint32                   // 獲得正在被使用的網頁下載器的數量。
}

// 建立網頁下載器池。
func NewPageDownloaderPool(
	total uint32,
	gen GenPageDownloader) (PageDownloaderPool, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() mdw.Entity {
		return gen()
	}
	pool, err := mdw.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myDownloaderPool{pool: pool, etype: etype}
	return dlpool, nil
}

// 網頁下載器池的實現型態。
type myDownloaderPool struct {
	pool  mdw.Pool     // 實體池。
	etype reflect.Type // 池內實體的型態。
}

func (dlpool *myDownloaderPool) Take() (PageDownloader, error) {
	entity, err := dlpool.pool.Take()
	if err != nil {
		return nil, err
	}
	dl, ok := entity.(PageDownloader)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", dlpool.etype)
		panic(errors.New(errMsg))
	}
	return dl, nil
}

func (dlpool *myDownloaderPool) Return(dl PageDownloader) error {
	return dlpool.pool.Return(dl)
}

func (dlpool *myDownloaderPool) Total() uint32 {
	return dlpool.pool.Total()
}
func (dlpool *myDownloaderPool) Used() uint32 {
	return dlpool.pool.Used()
}
