package base

import (
	"errors"
	"fmt"
)

// 參數容器的接口。
type Args interface {
	// 自檢參數的有效性，並在必要時傳回可以說明問題的錯誤值。
	// 若結果值為nil，則說明沒有找到問題，否則就表示自檢未透過。
	Check() error
	// 獲得參數容器的字串表現形式。
	String() string
}

// 通道參數的容器的描述範本。
var channelArgsTemplate string = "{ reqChanLen: %d, respChanLen: %d," +
	" itemChanLen: %d, errorChanLen: %d }"

// 通道參數的容器。
type ChannelArgs struct {
	reqChanLen   uint   // 請求通道的長度。
	respChanLen  uint   // 響應通道的長度。
	itemChanLen  uint   // 項目通道的長度。
	errorChanLen uint   // 錯誤通道的長度。
	description  string // 描述。
}

// 建立通道參數的容器。
func NewChannelArgs(
	reqChanLen uint,
	respChanLen uint,
	itemChanLen uint,
	errorChanLen uint) ChannelArgs {
	return ChannelArgs{
		reqChanLen:   reqChanLen,
		respChanLen:  respChanLen,
		itemChanLen:  itemChanLen,
		errorChanLen: errorChanLen,
	}
}

func (args *ChannelArgs) Check() error {
	if args.reqChanLen == 0 {
		return errors.New("The request channel max length (capacity) can not be 0!\n")
	}
	if args.respChanLen == 0 {
		return errors.New("The response channel max length (capacity) can not be 0!\n")
	}
	if args.itemChanLen == 0 {
		return errors.New("The item channel max length (capacity) can not be 0!\n")
	}
	if args.errorChanLen == 0 {
		return errors.New("The error channel max length (capacity) can not be 0!\n")
	}
	return nil
}

func (args *ChannelArgs) String() string {
	if args.description == "" {
		args.description =
			fmt.Sprintf(channelArgsTemplate,
				args.reqChanLen,
				args.respChanLen,
				args.itemChanLen,
				args.errorChanLen)
	}
	return args.description
}

// 獲得請求通道的長度。
func (args *ChannelArgs) ReqChanLen() uint {
	return args.reqChanLen
}

// 獲得響應通道的長度。
func (args *ChannelArgs) RespChanLen() uint {
	return args.respChanLen
}

// 獲得項目通道的長度。
func (args *ChannelArgs) ItemChanLen() uint {
	return args.itemChanLen
}

// 獲得錯誤通道的長度。
func (args *ChannelArgs) ErrorChanLen() uint {
	return args.errorChanLen
}

// 池基本參數容器的描述範本。
var poolBaseArgsTemplate string = "{ pageDownloaderPoolSize: %d," +
	" analyzerPoolSize: %d }"

// 池基本參數的容器。
type PoolBaseArgs struct {
	pageDownloaderPoolSize uint32 // 網頁下載器池的尺寸。
	analyzerPoolSize       uint32 // 分析器池的尺寸。
	description            string // 描述。
}

// 建立池基本參數的容器。
func NewPoolBaseArgs(
	pageDownloaderPoolSize uint32,
	analyzerPoolSize uint32) PoolBaseArgs {
	return PoolBaseArgs{
		pageDownloaderPoolSize: pageDownloaderPoolSize,
		analyzerPoolSize:       analyzerPoolSize,
	}
}

func (args *PoolBaseArgs) Check() error {
	if args.pageDownloaderPoolSize == 0 {
		return errors.New("The page downloader pool size can not be 0!\n")
	}
	if args.analyzerPoolSize == 0 {
		return errors.New("The analyzer pool size can not be 0!\n")
	}
	return nil
}

func (args *PoolBaseArgs) String() string {
	if args.description == "" {
		args.description =
			fmt.Sprintf(poolBaseArgsTemplate,
				args.pageDownloaderPoolSize,
				args.analyzerPoolSize)
	}
	return args.description
}

// 獲得網頁下載器池的尺寸。
func (args *PoolBaseArgs) PageDownloaderPoolSize() uint32 {
	return args.pageDownloaderPoolSize
}

// 獲得分析器池的尺寸。
func (args *PoolBaseArgs) AnalyzerPoolSize() uint32 {
	return args.analyzerPoolSize
}
