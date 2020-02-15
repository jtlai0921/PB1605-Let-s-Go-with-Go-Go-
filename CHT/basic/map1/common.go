package map1

import (
	"reflect"
)

// 泛化的Map的接口型態
type GenericMap interface {
	// 取得指定鍵值對應的元素值。若沒有對應元素值則傳回nil。
	Get(key interface{}) interface{}
	// 加入鍵值對，並傳回與指定鍵值對應的舊的元素值。若沒有舊元素值則傳回(nil, true)。
	Put(key interface{}, elem interface{}) (interface{}, bool)
	// 移除與指定鍵值對應的鍵值對，並傳回舊的元素值。若沒有舊元素值則傳回nil。
	Remove(key interface{}) interface{}
	// 清除所有的鍵值對。
	Clear()
	// 取得鍵值對的數量。
	Len() int
	// 判斷是否包括指定的鍵值。
	Contains(key interface{}) bool
	// 取得已排序的鍵值所群組成的切片值。
	Keys() []interface{}
	// 取得已排序的元素值所群組成的切片值。
	Elems() []interface{}
	// 取得已包括的鍵值對所群組成的字典值。
	ToMap() map[interface{}]interface{}
	// 取得鍵的型態。
	KeyType() reflect.Type
	// 取得元素的型態。
	ElemType() reflect.Type
}
