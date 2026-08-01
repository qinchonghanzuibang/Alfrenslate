package providers

import "sort"

func Ordered(order []string, available []Provider) []Provider {
	index := map[string]int{}
	for i, id := range order {
		index[id] = i
	}
	sort.SliceStable(available, func(i, j int) bool {
		a, aok := index[available[i].ID()]
		b, bok := index[available[j].ID()]
		if !aok {
			a = 1000
		}
		if !bok {
			b = 1000
		}
		return a < b
	})
	return available
}
