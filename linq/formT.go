package linq

// 筛选子元素字段
type linqFormT[T1 any, T2 any] struct {
	// source array
	source []T1
}

// FromT 筛选子元素字段
func FromT[T1 any, T2 any](source []T1) linqFormT[T1, T2] {
	return linqFormT[T1, T2]{
		source: source,
	}
}

// Select 筛选子元素字段
func (receiver linqFormT[T1, T2]) Select(fn func(item T1) T2) []T2 {
	var lst []T2
	for _, item := range receiver.source {
		lst = append(lst, fn(item))
	}
	return lst
}
