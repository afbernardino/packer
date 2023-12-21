package order

import "packer/internal/rest/order/pack"

type Request struct {
	Size int `json:"size"`
}

type Order struct {
	Packs []pack.Pack `json:"packs"`
}

type Config struct {
	PackSizes []int `json:"pack_sizes"`
}
