package pack

import (
	"maps"
	"slices"
)

type Computer struct{}

func NewComputer() Computer {
	return Computer{}
}

// ComputePacks the packs that need to be shipped to the customer given the pack sizes and the order size.
func (comp *Computer) ComputePacks(packSizes []int, orderSize int) []Pack {
	if !comp.isValidInput(packSizes, orderSize) {
		return []Pack{}
	}
	sortedPackSizes := comp.cloneAndSort(packSizes)
	return comp.compute(sortedPackSizes, orderSize)
}

func (comp *Computer) isValidInput(packSizes []int, orderSize int) bool {
	return len(packSizes) > 0 && orderSize > 0
}

func (comp *Computer) cloneAndSort(s []int) []int {
	cloned := slices.Clone(s)
	slices.Sort(cloned)
	return cloned
}

func (comp *Computer) compute(packSizes []int, orderSize int) []Pack {
	rawPacks := comp.computeRawPacks(packSizes, orderSize)
	optimizedPacks := comp.optimizePacks(packSizes, rawPacks)
	return comp.toPackModelSlice(optimizedPacks)
}

// computeRawPacks computes a slice with all the packs that can fulfill the order size.
// Some packs will be unoptimized
// (e.g. for pack sizes of 250 and 500 an order size of 251 it returns 2 packs of 250, instead of 1 pack of 500).
// See optimizePacks to optimize these packs.
func (comp *Computer) computeRawPacks(packSizes []int, orderSize int) []int {
	numPacks := make([]int, orderSize+1)
	usedPackSizes := make([]int, orderSize+1)

	placeholder := orderSize + 1
	numPacks = comp.fillWithPlaceholderFromIndex(numPacks, placeholder, 1)

	smallestPackSize := packSizes[0]
	usedPackSizes = comp.fillWithPlaceholderFromIndex(usedPackSizes, smallestPackSize, 1)

	for targetOrderSize := 1; targetOrderSize < len(numPacks); targetOrderSize++ {
		for _, packSize := range packSizes {
			if targetOrderSize-packSize < 0 {
				break
			}

			if numPacks[targetOrderSize-packSize]+1 < numPacks[targetOrderSize] {
				numPacks[targetOrderSize] = numPacks[targetOrderSize-packSize] + 1
				usedPackSizes[targetOrderSize] = packSize
			}
		}
	}

	return comp.computePacksByUsedPackSizes(usedPackSizes)
}

func (comp *Computer) fillWithPlaceholderFromIndex(numPacks []int, placeholder, index int) []int {
	for i := index; i < len(numPacks); i++ {
		numPacks[i] = placeholder
	}
	return numPacks
}

func (comp *Computer) computePacksByUsedPackSizes(usedPackSizes []int) []int {
	var result []int
	targetOrderSize := len(usedPackSizes) - 1
	for targetOrderSize > 0 {
		result = append(result, usedPackSizes[targetOrderSize])
		targetOrderSize -= usedPackSizes[targetOrderSize]
	}
	return result
}

// optimizePacks merges smaller packs into larger ones (e.g. 2 packs of 250 will be merged into one of 500, if it exists).
func (comp *Computer) optimizePacks(packSizes, packs []int) []int {
	quantityByPack := comp.packSliceToQuantityByPackMap(packs)
	mergedQuantityByPack := comp.mergeToLargerPacks(packSizes, quantityByPack)
	return comp.quantityByPackMapToPackSlice(mergedQuantityByPack)
}

func (comp *Computer) packSliceToQuantityByPackMap(packs []int) map[int]int {
	result := make(map[int]int)
	for _, pack := range packs {
		result[pack]++
	}
	return result
}

func (comp *Computer) mergeToLargerPacks(packSizes []int, quantityByPack map[int]int) map[int]int {
	for i := len(packSizes) - 1; i > 0; i-- {
		quantityByPack = comp.mergeToPackSize(packSizes[i], quantityByPack)
	}
	return quantityByPack
}

func (comp *Computer) mergeToPackSize(packSize int, quantityByPack map[int]int) map[int]int {
	result := maps.Clone(quantityByPack)
	for pack, quantity := range quantityByPack {
		sum := 0
		count := 0
		for i := 0; i < quantity; i++ {
			sum += pack
			count++
			if sum == packSize {
				result[pack] -= count
				result[packSize]++
				sum = 0
				count = 0
			}
		}
	}
	return result
}

func (comp *Computer) quantityByPackMapToPackSlice(quantityByPack map[int]int) []int {
	var result []int
	for pack, quantity := range quantityByPack {
		for i := 0; i < quantity; i++ {
			result = append(result, pack)
		}
	}
	return result
}

func (comp *Computer) toPackModelSlice(packs []int) []Pack {
	if len(packs) == 0 {
		return []Pack{}
	}

	quantityByPack := comp.packSliceToQuantityByPackMap(packs)

	var result []Pack
	for size, quantity := range quantityByPack {
		if quantity > 0 {
			result = append(result, Pack{
				Size:     size,
				Quantity: quantity,
			})
		}
	}
	return result
}
