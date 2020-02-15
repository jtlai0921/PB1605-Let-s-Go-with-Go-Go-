package base

import "logging"

// 建立日志記錄器。
func NewLogger() logging.Logger {
	return logging.NewSimpleLogger()
}
