package pack

// SizesValid return true if there are pack sizes and all are greater than zero, false otherwise.
func SizesValid(packSizes []int) bool {
	if packSizes == nil || len(packSizes) == 0 {
		return false
	}

	for _, packSize := range packSizes {
		if packSize <= 0 {
			return false
		}
	}

	return true
}

func RemoveDuplicateSizes(packSizes []int) []int {
	packSizesSet := sliceToSet(packSizes)
	return setToSlice(packSizesSet)
}

func sliceToSet(packSizes []int) map[int]bool {
	result := make(map[int]bool)
	for _, packSize := range packSizes {
		result[packSize] = true
	}
	return result
}

func setToSlice(packSizes map[int]bool) []int {
	var result []int
	for packSize := range packSizes {
		result = append(result, packSize)
	}
	return result
}

// EqualSlice returns true if both pack slices have the exact same packs (even if unordered), false otherwise.
func EqualSlice(s1 []Pack, s2 []Pack) bool {
	if len(s1) != len(s2) {
		return false
	}

	for _, p1 := range s1 {
		found := false
		for _, p2 := range s2 {
			if p1.Size == p2.Size && p1.Quantity == p2.Quantity {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
