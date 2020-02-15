package middleware

import (
	"math"
	"sync"
)

// ID產生器的接口型態。
type IdGenerator interface {
	GetUint32() uint32 // 獲得一個uint32型態的ID。
}

// 建立ID產生器。
func NewIdGenerator() IdGenerator {
	return &cyclicIdGenerator{}
}

// ID產生器的實現型態。
type cyclicIdGenerator struct {
	sn    uint32     // 目前的ID。
	ended bool       // 前一個ID是否已經為其型態所能表示的最大值。
	mutex sync.Mutex // 互斥鎖。
}

func (gen *cyclicIdGenerator) GetUint32() uint32 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	if gen.ended {
		defer func() { gen.ended = false }()
		gen.sn = 0
		return gen.sn
	}
	id := gen.sn
	if id < math.MaxUint32 {
		gen.sn++
	} else {
		gen.ended = true
	}
	return id
}

// ID產生器的接口型態2。
type IdGenerator2 interface {
	GetUint64() uint64 // 獲得一個uint64型態的ID。
}

// 建立ID產生器2。
func NewIdGenerator2() IdGenerator2 {
	return &cyclicIdGenerator2{}
}

// ID產生器的實現型態2。
type cyclicIdGenerator2 struct {
	base       cyclicIdGenerator // 基本的ID產生器。
	cycleCount uint64            // 基於uint32型態的取值範圍的周期計數。
}

func (gen *cyclicIdGenerator2) GetUint64() uint64 {
	var id64 uint64
	if gen.cycleCount%2 == 1 {
		id64 += math.MaxUint32
	}
	id32 := gen.base.GetUint32()
	if id32 == math.MaxUint32 {
		gen.cycleCount++
	}
	id64 += uint64(id32)
	return id64
}
