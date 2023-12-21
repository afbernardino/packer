package pack

import (
	"fmt"
	"slices"
	"testing"
)

func TestComputer_Compute_Success(t *testing.T) {
	data := []struct {
		packSizes     []int
		orderSize     int
		expectedPacks []Pack
	}{
		{
			packSizes:     []int{250, 500},
			orderSize:     -1,
			expectedPacks: []Pack{},
		},
		{
			packSizes:     []int{250, 500},
			orderSize:     0,
			expectedPacks: []Pack{},
		},
		{
			packSizes: []int{250, 500, 1000, 2000, 5000},
			orderSize: 1,
			expectedPacks: []Pack{
				{
					Size:     250,
					Quantity: 1,
				},
			},
		},
		{
			packSizes: []int{250, 500, 1000, 2000, 5000},
			orderSize: 250,
			expectedPacks: []Pack{
				{
					Size:     250,
					Quantity: 1,
				},
			},
		},
		{
			packSizes: []int{250, 500, 1000, 2000, 5000},
			orderSize: 251,
			expectedPacks: []Pack{
				{
					Size:     500,
					Quantity: 1,
				},
			},
		},
		{
			packSizes: []int{250, 500, 1000, 2000, 5000},
			orderSize: 501,
			expectedPacks: []Pack{
				{
					Size:     500,
					Quantity: 1,
				},
				{
					Size:     250,
					Quantity: 1,
				},
			},
		},
		{
			packSizes: []int{250, 500, 1000, 2000, 5000},
			orderSize: 12001,
			expectedPacks: []Pack{
				{
					Size:     5000,
					Quantity: 2,
				},
				{
					Size:     2000,
					Quantity: 1,
				},
				{
					Size:     250,
					Quantity: 1,
				},
			},
		},
		{
			packSizes: []int{23, 31, 53},
			orderSize: 500000,
			expectedPacks: []Pack{
				{
					Size:     53,
					Quantity: 9429,
				},
				{
					Size:     31,
					Quantity: 7,
				},
				{
					Size:     23,
					Quantity: 2,
				},
			},
		},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf("with order size: %d", d.orderSize), func(t *testing.T) {
			comp := NewComputer()

			packs := comp.ComputePacks(d.packSizes, d.orderSize)

			if !EqualSlice(packs, d.expectedPacks) {
				t.Errorf("unexpected packs: got '%+v' want '%+v'", packs, d.expectedPacks)
			}
		})
	}
}

func TestComputer_Compute_UnorderedPackSizes(t *testing.T) {
	comp := NewComputer()
	packSizes := []int{250, 2000, 1000, 500, 5000}

	packs := comp.ComputePacks(packSizes, 12001)

	expectedPacks := []Pack{
		{
			Size:     5000,
			Quantity: 2,
		},
		{
			Size:     2000,
			Quantity: 1,
		},
		{
			Size:     250,
			Quantity: 1,
		},
	}
	if !EqualSlice(packs, expectedPacks) {
		t.Errorf("unexpected packs: got '%+v' want '%+v'", packs, expectedPacks)
	}
}

func TestComputer_Compute_PackSizesSliceNotAffected(t *testing.T) {
	comp := NewComputer()
	packSizes := []int{250, 2000, 1000, 500, 5000}
	oracle := []int{250, 2000, 1000, 500, 5000}

	packs := comp.ComputePacks(packSizes, 12001)

	if !slices.Equal(packSizes, oracle) {
		t.Errorf("unexpected packs: got '%+v' want '%+v'", packs, oracle)
	}
}
