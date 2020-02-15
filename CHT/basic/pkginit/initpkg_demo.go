package main // 指令原始程式檔案必須在這裡宣告自己屬於main包

import ( // 引入了程式碼包fmt和runtime
	"fmt"
	"runtime"
)

func init() { // 包起始化函數
	fmt.Printf("Map: %v\n", m) // 先格式化再列印
	// 透過呼叫runtime套件的程式碼取得目前機器所執行的動作系統以及計算架構
	// 而後透過fmt套件的Sprintf方法進行字串格式化並給予值給變數info
	info = fmt.Sprintf("OS: %s, Arch: %s", runtime.GOOS, runtime.GOARCH)
}

var m map[int]string = map[int]string{1: "A", 2: "B", 3: "C"} // 非局部變數，map型態，且已起始化

var info string // 非局部變數，string型態，未被起始化

func main() { // 指令原始程式檔案必須有的入口函數
	fmt.Println(info) // 列印變數info
}
