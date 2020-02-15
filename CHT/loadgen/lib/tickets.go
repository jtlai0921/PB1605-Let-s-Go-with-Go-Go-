package lib

import (
	"errors"
	"fmt"
)

// Goroutine票池的接口。
type GoTickets interface {
	// 拿走一張票。
	Take()
	// 歸還一張票。
	Return()
	// 票池是否已被啟動。
	Active() bool
	// 票的總數。
	Total() uint32
	// 剩余的票數。
	Remainder() uint32
}

// Goroutine票池的實現。
type myGoTickets struct {
	total    uint32    // 票的總數。
	ticketCh chan byte // 票的容器。
	active   bool      // 票池是否已被啟動。
}

func NewGoTickets(total uint32) (GoTickets, error) {
	gt := myGoTickets{}
	if !gt.init(total) {
		errMsg :=
			fmt.Sprintf("The goroutine ticket pool can NOT be initialized! (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	return &gt, nil
}

func (gt *myGoTickets) init(total uint32) bool {
	if gt.active {
		return false
	}
	if total == 0 {
		return false
	}
	ch := make(chan byte, total)
	n := int(total)
	for i := 0; i < n; i++ {
		ch <- 1
	}
	gt.ticketCh = ch
	gt.total = total
	gt.active = true
	return true
}

func (gt *myGoTickets) Take() {
	<-gt.ticketCh
}

func (gt *myGoTickets) Return() {
	gt.ticketCh <- 1
}

func (gt *myGoTickets) Active() bool {
	return gt.active
}

func (gt *myGoTickets) Total() uint32 {
	return gt.total
}

func (gt *myGoTickets) Remainder() uint32 {
	return uint32(len(gt.ticketCh))
}
