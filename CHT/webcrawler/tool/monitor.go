package tool

import (
	"errors"
	"fmt"
	"runtime"
	"time"
	sched "webcrawler/scheduler"
)

// 摘要訊息的範本。
var summaryForMonitoring = "Monitor - Collected information[%d]:\n" +
	"  Goroutine number: %d\n" +
	"  Scheduler:\n%s" +
	"  Escaped time: %s\n"

// 已達到最大閒置計數的訊息範本。
var msgReachMaxIdleCount = "The scheduler has been idle for a period of time" +
	" (about %s)." +
	" Now consider what stop it."

// 停止分派器的訊息範本。
var msgStopScheduler = "Stop scheduler...%s."

// 日志記錄函數的型態。
// 參數level代表日志等級。等級設定：0：普通；1：警示；2：錯誤。
type Record func(level byte, content string)

// 分派器監控函數。
// 參數scheduler代表作為監控目的的分派器。
// 參數intervalNs代表檢查間隔時間，單位：毫微秒。
// 參數maxIdleCount代表最大閒置計數。
// 參數autoStop被用來指示該方法是否在分派器閒置一段時間（即持續閒置時間，由intervalNs * maxIdleCount得出）之後自行停止分派器。
// 參數detailSummary被用來表示是否需要詳細的摘要訊息。
// 參數record代表日志記錄函數。
// 當監控結束之後，該方法會會向作為唯一傳回值的通道傳送一個代表了閒置狀態檢查次數的數值。
func Monitoring(
	scheduler sched.Scheduler,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	detailSummary bool,
	record Record) <-chan uint64 {
	if scheduler == nil { // 分派器不能不可用！
		panic(errors.New("The scheduler is invalid!"))
	}
	// 防止過小的參數值對爬取流程的影響
	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}
	if maxIdleCount < 1000 {
		maxIdleCount = 1000
	}
	// 監控停止知會器
	stopNotifier := make(chan byte, 1)
	// 接收和報告錯誤
	reportError(scheduler, record, stopNotifier)
	// 記錄摘要訊息
	recordSummary(scheduler, detailSummary, record, stopNotifier)
	// 檢查計數通道
	checkCountChan := make(chan uint64, 2)
	// 檢查閒置狀態
	checkStatus(scheduler,
		intervalNs,
		maxIdleCount,
		autoStop,
		checkCountChan,
		record,
		stopNotifier)
	return checkCountChan
}

// 檢查狀態，並在滿足持續閒置時間的條件時采取必要措施。
func checkStatus(
	scheduler sched.Scheduler,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	checkCountChan chan<- uint64,
	record Record,
	stopNotifier chan<- byte) {
	var checkCount uint64
	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			checkCountChan <- checkCount
		}()
		// 等待分派器開啟
		waitForSchedulerStart(scheduler)
		// 準備
		var idleCount uint
		var firstIdleTime time.Time
		for {
			// 檢查分派器的閒置狀態
			if scheduler.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}
				if idleCount >= maxIdleCount {
					msg :=
						fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String())
					record(0, msg)
					// 再次檢查分派器的閒置狀態，確保它已經可以被停止
					if scheduler.Idle() {
						if autoStop {
							var result string
							if scheduler.Stop() {
								result = "success"
							} else {
								result = "failing"
							}
							msg = fmt.Sprintf(msgStopScheduler, result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(intervalNs)
		}
	}()
}

// 記錄摘要訊息。
func recordSummary(
	scheduler sched.Scheduler,
	detailSummary bool,
	record Record,
	stopNotifier <-chan byte) {
	go func() {
		// 等待分派器開啟
		waitForSchedulerStart(scheduler)
		// 準備
		var prevSchedSummary sched.SchedSummary
		var prevNumGoroutine int
		var recordCount uint64 = 1
		startTime := time.Now()
		for {
			// 檢視監控停止知會器
			select {
			case <-stopNotifier:
				return
			default:
			}
			// 取得摘要訊息的各群組成部分
			currNumGoroutine := runtime.NumGoroutine()
			currSchedSummary := scheduler.Summary("    ")
			// 比對前後兩份摘要訊息的一致性。只有不一致時才會予以記錄。
			if currNumGoroutine != prevNumGoroutine ||
				!currSchedSummary.Same(prevSchedSummary) {
				schedSummaryStr := func() string {
					if detailSummary {
						return currSchedSummary.Detail()
					} else {
						return currSchedSummary.String()
					}
				}()
				// 記錄摘要訊息
				info := fmt.Sprintf(summaryForMonitoring,
					recordCount,
					currNumGoroutine,
					schedSummaryStr,
					time.Since(startTime).String(),
				)
				record(0, info)
				prevNumGoroutine = currNumGoroutine
				prevSchedSummary = currSchedSummary
				recordCount++
			}
			time.Sleep(time.Microsecond)
		}
	}()
}

// 接收和報告錯誤。
func reportError(
	scheduler sched.Scheduler,
	record Record,
	stopNotifier <-chan byte) {
	go func() {
		// 等待分派器開啟
		waitForSchedulerStart(scheduler)
		for {
			// 檢視監控停止知會器
			select {
			case <-stopNotifier:
				return
			default:
			}
			errorChan := scheduler.ErrorChan()
			if errorChan == nil {
				return
			}
			err := <-errorChan
			if err != nil {
				errMsg := fmt.Sprintf("Error (received from error channel): %s", err)
				record(2, errMsg)
			}
			time.Sleep(time.Microsecond)
		}
	}()
}

// 等待分派器開啟。
func waitForSchedulerStart(scheduler sched.Scheduler) {
	for !scheduler.Running() {
		time.Sleep(time.Microsecond)
	}
}
