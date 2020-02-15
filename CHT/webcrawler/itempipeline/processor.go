package itemproc

import (
	base "webcrawler/base"
)

// 被用來處理項目的函數型態。
type ProcessItem func(item base.Item) (result base.Item, err error)
