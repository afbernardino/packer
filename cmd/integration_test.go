//go:build integration_test

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"packer/internal/rest/order"
	"packer/internal/rest/order/pack"
	"slices"
	"testing"
	"time"
)

const (
	url             = "http://localhost:8080"
	ordersUrl       = url + "/orders"
	ordersConfigUrl = ordersUrl + "/config"
)

func TestValidSetConfigAndCreateOrders(t *testing.T) {
	data := []struct {
		packSizes     []int
		orderSize     int
		expectedPacks []pack.Pack
	}{
		{
			packSizes: []int{250},
			orderSize: 250,
			expectedPacks: []pack.Pack{
				{
					Size:     250,
					Quantity: 1,
				},
			},
		},
		{
			packSizes: []int{250, 250, 500, 1000, 2000, 5000},
			orderSize: 12001,
			expectedPacks: []pack.Pack{
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
			expectedPacks: []pack.Pack{
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
			httpClient := newHttpClient()

			setConfigResp := doValidSetConfig(t, &httpClient, d.packSizes)
			assertValidSetConfig(t, setConfigResp, d.packSizes)

			createOrderResp := doValidCreateOrder(t, &httpClient, d.orderSize)
			assertValidCreateOrder(t, createOrderResp, d.expectedPacks)
		})
	}
}

func TestSetConfig_InvalidPayload(t *testing.T) {
	httpClient := newHttpClient()
	req := newPutRequest(t, ordersConfigUrl, []byte("{pack_sizes}"))

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertInvalidPayloadResponse(t, resp)
}

func TestSetConfig_InvalidPackSizes(t *testing.T) {
	httpClient := newHttpClient()

	cfgsWithInvalidPackSizes := []order.Config{
		{
			PackSizes: []int{},
		},
		{
			PackSizes: []int{-1, 250},
		},
		{
			PackSizes: []int{0, 250},
		},
	}

	for _, cfg := range cfgsWithInvalidPackSizes {
		req := newSetConfigRequestWithConfig(t, cfg)

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		assertInvalidPackSizes(t, resp)
	}
}

func TestSetConfig_CorsHeaders(t *testing.T) {
	httpClient := newHttpClient()
	req := newOptionsRequest(t, ordersConfigUrl)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	assertCorsHeaders(t, resp)
}

func TestCreateOrder_InvalidOrderSize(t *testing.T) {
	httpClient := newHttpClient()

	invalidSizes := []int{-1, 0}

	for _, size := range invalidSizes {
		t.Run(fmt.Sprintf("with size: %d", size), func(t *testing.T) {
			req := newCreateOrderRequestWithSize(t, size)

			resp, err := httpClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			assertInvalidOrderSize(t, resp)
		})
	}
}

func TestCreateOrder_InvalidPayload(t *testing.T) {
	httpClient := newHttpClient()
	req := newPostRequest(t, ordersUrl, []byte("{size}"))

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	assertInvalidPayloadResponse(t, resp)
}

func TestCreateOrder_CorsHeaders(t *testing.T) {
	httpClient := newHttpClient()
	req := newOptionsRequest(t, ordersUrl)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	assertCorsHeaders(t, resp)
}

func newHttpClient() http.Client {
	return http.Client{Timeout: 30 * time.Second}
}

func newPutRequest(t *testing.T, url string, payload []byte) *http.Request {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func newPostRequest(t *testing.T, url string, payload []byte) *http.Request {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func newOptionsRequest(t *testing.T, url string) *http.Request {
	req, err := http.NewRequest(http.MethodOptions, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func newSetConfigRequestWithConfig(t *testing.T, cfg order.Config) *http.Request {
	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	return newPutRequest(t, ordersConfigUrl, jsonBytes)
}

func newCreateOrderRequestWithSize(t *testing.T, orderSize int) *http.Request {
	orderReq := order.Request{
		Size: orderSize,
	}
	jsonBytes, err := json.Marshal(orderReq)
	if err != nil {
		t.Fatal(err)
	}

	return newPostRequest(t, ordersUrl, jsonBytes)
}

func doValidSetConfig(t *testing.T, httpClient *http.Client, packSizes []int) *http.Response {
	cfg := order.Config{
		PackSizes: packSizes,
	}
	req := newSetConfigRequestWithConfig(t, cfg)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return resp
}

func assertValidSetConfig(t *testing.T, resp *http.Response, packSizes []int) {
	defer resp.Body.Close()

	var cfgResp order.Config
	if err := json.NewDecoder(resp.Body).Decode(&cfgResp); err != nil {
		t.Fatal(err)
	}

	slices.Sort(cfgResp.PackSizes)
	slices.Sort(packSizes)
	if !slices.Equal(cfgResp.PackSizes, packSizes) {
		t.Errorf("unexpected pack sizes: got '%+v' want '%+v'", cfgResp.PackSizes, packSizes)
	}

	assertHeader(t, resp, "Content-Type", "application/json")
}

func doValidCreateOrder(t *testing.T, httpClient *http.Client, orderSize int) *http.Response {
	req := newCreateOrderRequestWithSize(t, orderSize)
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func assertValidCreateOrder(t *testing.T, resp *http.Response, expectedPacks []pack.Pack) {
	defer resp.Body.Close()

	var orderResp order.Order
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		t.Fatal(err)
	}

	if !pack.EqualSlice(orderResp.Packs, expectedPacks) {
		t.Errorf("unexpected packs: got '%+v' want '%+v'", orderResp.Packs, expectedPacks)
	}

	assertHeader(t, resp, "Content-Type", "application/json")
}

func assertInvalidPayloadResponse(t *testing.T, resp *http.Response) {
	defer resp.Body.Close()

	var errorResponse order.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		t.Fatal(err)
	}

	expectedErrRespCode := "invalid_payload"
	if errorResponse.Code != expectedErrRespCode {
		t.Errorf("unexpected packs: got '%s' want '%s'", errorResponse.Code, expectedErrRespCode)
	}

	expectedErrRespMsg := "Invalid payload."
	if errorResponse.Message != expectedErrRespMsg {
		t.Errorf("unexpected packs: got '%s' want '%s'", errorResponse.Code, expectedErrRespMsg)
	}
}

func assertInvalidPackSizes(t *testing.T, resp *http.Response) {
	defer resp.Body.Close()

	var errorResponse order.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		t.Fatal(err)
	}

	expectedErrRespCode := "invalid_pack_sizes"
	if errorResponse.Code != expectedErrRespCode {
		t.Errorf("unexpected packs: got '%s' want '%s'", errorResponse.Code, expectedErrRespCode)
	}

	expectedErrRespMsg := "Pack sizes should have at least one size and all sizes should be greater than zero."
	if errorResponse.Message != expectedErrRespMsg {
		t.Errorf("unexpected packs: got '%s' want '%s'", errorResponse.Message, expectedErrRespMsg)
	}
}

func assertInvalidOrderSize(t *testing.T, resp *http.Response) {
	defer resp.Body.Close()

	var errorResponse order.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		t.Fatal(err)
	}

	expectedErrRespCode := "invalid_order_size"
	if errorResponse.Code != expectedErrRespCode {
		t.Errorf("unexpected packs: got '%s' want '%s'", errorResponse.Code, expectedErrRespCode)
	}

	expectedErrRespMsg := "Order sizes must be greater than zero."
	if errorResponse.Message != expectedErrRespMsg {
		t.Errorf("unexpected packs: got '%s' want '%s'", errorResponse.Code, expectedErrRespMsg)
	}
}

func assertCorsHeaders(t *testing.T, resp *http.Response) {
	assertHeader(t, resp, "Access-Control-Allow-Origin", "*")
	assertHeader(t, resp, "Access-Control-Allow-Methods", "POST, PUT")
	assertHeader(t, resp, "Access-Control-Allow-Headers", "*")
}

func assertHeader(t *testing.T, resp *http.Response, name, expected string) {
	headerValue := resp.Header.Get(name)
	if headerValue != expected {
		t.Errorf("unexpected http header: got '%s' want '%s'", headerValue, expected)
	}
}
