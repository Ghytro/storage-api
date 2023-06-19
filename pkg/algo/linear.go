package algo

import "golang.org/x/exp/constraints"

func Map[TSource, TDest any](collection []TSource, iteratee func(item TSource, idx int) TDest) []TDest {
	result := make([]TDest, 0, len(collection))
	for i, item := range collection {
		result = append(result, iteratee(item, i))
	}
	return result
}

func FlatMap[TSource, TDest any, TDestSlice []TDest](collection []TSource, iteratee func(item TSource, idx int) []TDest) []TDest {
	result := []TDest{}
	for i, item := range collection {
		result = append(result, iteratee(item, i)...)
	}
	return result
}

func SumBy[T any, T1 constraints.Ordered](collection []T, iteratee func(item T, idx int) T1) T1 {
	var result T1
	for i, item := range collection {
		result += iteratee(item, i)
	}
	return result
}

func Find[T any](collection []T, iteratee func(item T) bool) (T, bool) {
	for _, el := range collection {
		if iteratee(el) {
			return el, true
		}
	}
	var t T
	return t, false
}

func Filter[T any](collection []T, iteratee func(item T, idx int) bool) []T {
	result := []T{}
	for i, el := range collection {
		if iteratee(el, i) {
			result = append(result, el)
		}
	}
	return result
}

func UniqBy[T any, T1 comparable](collection []T, iteratee func(item T) T1) []T {
	m := map[T1]T{}
	for _, el := range collection {
		m[iteratee(el)] = el
	}
	result := make([]T, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
