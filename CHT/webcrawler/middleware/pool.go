package middleware

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// 實體的接口型態。
type Entity interface {
	Id() uint32 // ID的取得方法。
}

// 實體池的接口型態。
type Pool interface {
	Take() (Entity, error)      // 取出實體
	Return(entity Entity) error // 歸還實體。
	Total() uint32              // 實體池的容量。
	Used() uint32               // 實體池中已被使用的實體的數量。
}

// 建立實體池。
func NewPool(
	total uint32,
	entityType reflect.Type,
	genEntity func() Entity) (Pool, error) {
	if total == 0 {
		errMsg :=
			fmt.Sprintf("The pool can not be initialized! (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	size := int(total)
	container := make(chan Entity, size)
	idContainer := make(map[uint32]bool)
	for i := 0; i < size; i++ {
		newEntity := genEntity()
		if entityType != reflect.TypeOf(newEntity) {
			errMsg :=
				fmt.Sprintf("The type of result of function genEntity() is NOT %s!\n", entityType)
			return nil, errors.New(errMsg)
		}
		container <- newEntity
		idContainer[newEntity.Id()] = true
	}
	pool := &myPool{
		total:       total,
		etype:       entityType,
		genEntity:   genEntity,
		container:   container,
		idContainer: idContainer,
	}
	return pool, nil
}

// 實體池的實現型態。
type myPool struct {
	total       uint32          // 池的總容量。
	etype       reflect.Type    // 池中實體的型態。
	genEntity   func() Entity   // 池中實體的產生函數。
	container   chan Entity     // 實體容器。
	idContainer map[uint32]bool // 實體ID的容器。
	mutex       sync.Mutex      // 針對實體ID容器動作的互斥鎖。
}

func (pool *myPool) Take() (Entity, error) {
	entity, ok := <-pool.container
	if !ok {
		return nil, errors.New("The inner container is invalid!")
	}
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.idContainer[entity.Id()] = false
	return entity, nil
}

func (pool *myPool) Return(entity Entity) error {
	if entity == nil {
		return errors.New("The returning entity is invalid!")
	}
	if pool.etype != reflect.TypeOf(entity) {
		errMsg := fmt.Sprintf("The type of returning entity is NOT %s!\n", pool.etype)
		return errors.New(errMsg)
	}
	entityId := entity.Id()
	casResult := pool.compareAndSetForIdContainer(entityId, false, true)
	if casResult == 1 {
		pool.container <- entity
		return nil
	} else if casResult == 0 {
		errMsg := fmt.Sprintf("The entity (id=%d) is already in the pool!\n", entityId)
		return errors.New(errMsg)
	} else {
		errMsg := fmt.Sprintf("The entity (id=%d) is illegal!\n", entityId)
		return errors.New(errMsg)
	}
}

// 比較並設定實體ID容器中與指定實體ID對應的鍵值對的元素值。
// 結果值：
//       -1：表示鍵值對不存在。
//        0：表示動作失敗。
//        1：表示動作成功。
func (pool *myPool) compareAndSetForIdContainer(
	entityId uint32, oldValue bool, newValue bool) int8 {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	v, ok := pool.idContainer[entityId]
	if !ok {
		return -1
	}
	if v != oldValue {
		return 0
	}
	pool.idContainer[entityId] = newValue
	return 1
}

func (pool *myPool) Total() uint32 {
	return pool.total
}

func (pool *myPool) Used() uint32 {
	return pool.total - uint32(len(pool.container))
}
