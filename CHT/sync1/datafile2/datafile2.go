package datafile2

import (
	"errors"
	"io"
	"os"
	"sync"
)

// 資料的型態
type Data []byte

// 資料檔的接口型態。
type DataFile interface {
	// 讀取一個資料區塊。
	Read() (rsn int64, d Data, err error)
	// 寫入一個資料區塊。
	Write(d Data) (wsn int64, err error)
	// 取得最後讀取的資料區塊的序號。
	Rsn() int64
	// 取得最後寫入的資料區塊的序號。
	Wsn() int64
	// 取得資料區塊的長度
	DataLen() uint32
}

// 資料檔的實現型態。
type myDataFile struct {
	f       *os.File     // 檔案。
	fmutex  sync.RWMutex // 被用於檔案的讀寫鎖。
	rcond   *sync.Cond   //讀取操作需要用到的條件變數
	woffset int64        // 寫入操作需要用到的偏移量。
	roffset int64        // 讀取操作需要用到的偏移量。
	wmutex  sync.Mutex   // 寫入操作需要用到的互斥鎖。
	rmutex  sync.Mutex   // 讀取操作需要用到的互斥鎖。
	dataLen uint32       // 資料區塊長度。
}

func NewDataFile(path string, dataLen uint32) (DataFile, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if dataLen == 0 {
		return nil, errors.New("Invalid data length!")
	}
	df := &myDataFile{f: f, dataLen: dataLen}
	df.rcond = sync.NewCond(df.fmutex.RLocker())
	return df, nil
}

func (df *myDataFile) Read() (rsn int64, d Data, err error) {
	// 讀取並更新讀偏移量
	var offset int64
	df.rmutex.Lock()
	offset = df.roffset
	df.roffset += int64(df.dataLen)
	df.rmutex.Unlock()

	//讀取一個資料區塊
	rsn = offset / int64(df.dataLen)
	bytes := make([]byte, df.dataLen)
	df.fmutex.RLock()
	defer df.fmutex.RUnlock()
	for {
		_, err = df.f.ReadAt(bytes, offset)
		if err != nil {
			if err == io.EOF {
				df.rcond.Wait()
				continue
			}
			return
		}
		d = bytes
		return
	}
}

func (df *myDataFile) Write(d Data) (wsn int64, err error) {
	// 讀取並更新寫偏移量
	var offset int64
	df.wmutex.Lock()
	offset = df.woffset
	df.woffset += int64(df.dataLen)
	df.wmutex.Unlock()

	//寫入一個資料區塊
	wsn = offset / int64(df.dataLen)
	var bytes []byte
	if len(d) > int(df.dataLen) {
		bytes = d[0:df.dataLen]
	} else {
		bytes = d
	}
	df.fmutex.Lock()
	defer df.fmutex.Unlock()
	_, err = df.f.Write(bytes)
	df.rcond.Signal()
	return
}

func (df *myDataFile) Rsn() int64 {
	df.rmutex.Lock()
	defer df.rmutex.Unlock()
	return df.roffset / int64(df.dataLen)
}

func (df *myDataFile) Wsn() int64 {
	df.wmutex.Lock()
	defer df.wmutex.Unlock()
	return df.woffset / int64(df.dataLen)
}

func (df *myDataFile) DataLen() uint32 {
	return df.dataLen
}
