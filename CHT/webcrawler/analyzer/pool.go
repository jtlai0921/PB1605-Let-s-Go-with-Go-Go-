package analyzer

import (
	"errors"
	"fmt"
	"reflect"
	mdw "webcrawler/middleware"
)

// 產生分析器的函數型態。
type GenAnalyzer func() Analyzer

// 分析器池的接口型態。
type AnalyzerPool interface {
	Take() (Analyzer, error)        // 從池中取出一個分析器。
	Return(analyzer Analyzer) error // 把一個分析器歸還給池。
	Total() uint32                  // 獲得池的總容量。
	Used() uint32                   // 獲得正在被使用的分析器的數量。
}

func NewAnalyzerPool(
	total uint32,
	gen GenAnalyzer) (AnalyzerPool, error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() mdw.Entity {
		return gen()
	}
	pool, err := mdw.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myAnalyzerPool{pool: pool, etype: etype}
	return dlpool, nil
}

type myAnalyzerPool struct {
	pool  mdw.Pool     // 實體池。
	etype reflect.Type // 池內實體的型態。
}

func (spdpool *myAnalyzerPool) Take() (Analyzer, error) {
	entity, err := spdpool.pool.Take()
	if err != nil {
		return nil, err
	}
	analyzer, ok := entity.(Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", spdpool.etype)
		panic(errors.New(errMsg))
	}
	return analyzer, nil
}

func (spdpool *myAnalyzerPool) Return(analyzer Analyzer) error {
	return spdpool.pool.Return(analyzer)
}

func (spdpool *myAnalyzerPool) Total() uint32 {
	return spdpool.pool.Total()
}
func (spdpool *myAnalyzerPool) Used() uint32 {
	return spdpool.pool.Used()
}
