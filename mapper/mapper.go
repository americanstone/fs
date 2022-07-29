package mapper

import (
	"github.com/devfeel/mapper"
)

// Array 数组转换
func Array[T any](fromSlice any) []T {
	var toSlice []T
	_ = mapper.MapperSlice(fromSlice, &toSlice)
	return toSlice
}

// Single 单个转换
func Single[T any](fromObj any) T {
	var toObj T
	_ = mapper.MapperSlice(fromObj, &toObj)
	return toObj
}
