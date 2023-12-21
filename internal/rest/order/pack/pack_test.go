package pack

import (
	"fmt"
	"slices"
	"testing"
)

func TestSizesValid_NilPackSizes(t *testing.T) {
	ok := SizesValid(nil)

	if ok {
		t.Error("pack sizes should be invalid")
	}
}

func TestSizesValid_EmptyPackSizes(t *testing.T) {
	ok := SizesValid([]int{})

	if ok {
		t.Error("pack sizes should be invalid")
	}
}

func TestSizesValid_InvalidPackSizes(t *testing.T) {
	invalidPackSizes := []int{-1, 0}

	for _, packSize := range invalidPackSizes {
		t.Run(fmt.Sprintf("with pack size: %d", packSize), func(t *testing.T) {
			ok := SizesValid([]int{packSize})

			if ok {
				t.Error("pack sizes should be invalid")
			}
		})
	}
}

func TestSizesValid_ValidPackSizes(t *testing.T) {
	ok := SizesValid([]int{250, 500})

	if !ok {
		t.Error("pack sizes should be valid")
	}
}

func TestRemoveDuplicateSizes(t *testing.T) {
	packSizes := []int{250, 250, 500, 500, 1000}

	result := RemoveDuplicateSizes(packSizes)

	slices.Sort(result)
	expectedResult := []int{250, 500, 1000}
	if !slices.Equal(result, expectedResult) {
		t.Errorf("unexpected packs: got '%+v' want '%+v'", result, expectedResult)
	}
}

func TestEqualSlice_Equal(t *testing.T) {
	s1 := []Pack{
		{
			Size:     500,
			Quantity: 2,
		},
		{
			Size:     250,
			Quantity: 1,
		},
	}

	s2 := []Pack{
		{
			Size:     250,
			Quantity: 1,
		},
		{
			Size:     500,
			Quantity: 2,
		},
	}

	ok := EqualSlice(s1, s2)

	if !ok {
		t.Error("slices should be equal")
	}
}

func TestEqualSlice_DifferentQuantities(t *testing.T) {
	s1 := []Pack{
		{
			Size:     500,
			Quantity: 2,
		},
		{
			Size:     250,
			Quantity: 1,
		},
	}

	s2 := []Pack{
		{
			Size:     500,
			Quantity: 1,
		},
		{
			Size:     250,
			Quantity: 1,
		},
	}

	ok := EqualSlice(s1, s2)

	if ok {
		t.Error("slices should not be equal")
	}
}

func TestEqualSlice_DifferentSizes(t *testing.T) {
	s1 := []Pack{
		{
			Size:     500,
			Quantity: 2,
		},
		{
			Size:     250,
			Quantity: 1,
		},
	}

	s2 := []Pack{
		{
			Size:     250,
			Quantity: 1,
		},
	}

	ok := EqualSlice(s1, s2)

	if ok {
		t.Error("slices should not be equal")
	}
}
