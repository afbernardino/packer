package order

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"packer/internal/rest/order/pack"
	"packer/internal/rest/order/repository"
	"slices"
	"testing"
)

type TestPackComputer struct {
	passedPackSizes []int
	passedOrderSize int
	result          []pack.Pack
}

func (comp *TestPackComputer) ComputePacks(packSizes []int, orderSize int) []pack.Pack {
	comp.passedPackSizes = packSizes
	comp.passedOrderSize = orderSize
	return comp.result
}

type TestErrRepository struct{}

func (_ *TestErrRepository) SetConfig(_ context.Context, _ repository.Config) error {
	return errors.New("test error")
}

func (_ *TestErrRepository) FindConfig(_ context.Context) (repository.Config, error) {
	return repository.Config{}, errors.New("test error")
}

type TestSuccessRepository struct {
	passedCfg repository.Config
	result    repository.Config
}

func (repo *TestSuccessRepository) SetConfig(_ context.Context, cfg repository.Config) error {
	repo.passedCfg = cfg
	return nil
}

func (repo *TestSuccessRepository) FindConfig(_ context.Context) (repository.Config, error) {
	return repo.result, nil
}

func TestServeHTTP_HandleCreateOrder_Success(t *testing.T) {
	data := []struct {
		computerResult []pack.Pack
	}{
		{
			computerResult: []pack.Pack{
				{
					Size:     5000,
					Quantity: 2,
				},
				{
					Size:     2000,
					Quantity: 1,
				}, {
					Size:     250,
					Quantity: 1,
				},
			},
		},
		{
			computerResult: []pack.Pack{},
		},
	}

	for _, d := range data {
		t.Run(fmt.Sprintf("with computer result %+v", d.computerResult), func(t *testing.T) {
			comp := TestPackComputer{
				result: d.computerResult,
			}
			cfg := repository.Config{
				PackSizes: []int{100, 200},
			}
			repo := TestSuccessRepository{result: cfg}
			handler := NewHandler(&comp, &repo)

			rr := httptest.NewRecorder()
			req := newCreateOrderRequestWithPayload(t, `{"size": 1}`)

			handler.ServeHTTP(rr, req)

			assertPackComputerReceivedPackSizes(t, comp, cfg.PackSizes)
			assertPackComputerReceivedOrderSize(t, comp, 1)
			assertStatusOk(t, rr)
			assertOrderPacks(t, rr, d.computerResult)
		})
	}
}

func TestServeHTTP_HandleCreateOrder_InvalidPayload(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateOrderRequestWithPayload(t, "{size}")

	handler.ServeHTTP(rr, req)

	assertBadRequestResponse(t, rr, "invalid_payload", "Invalid payload.")
}

func TestServeHTTP_HandleCreateOrder_InvalidOrderSize(t *testing.T) {
	invalidSizes := []int{-1, 0}

	for _, size := range invalidSizes {
		t.Run(fmt.Sprintf("with size: %d", size), func(t *testing.T) {
			comp := TestPackComputer{}
			repo := TestSuccessRepository{}
			handler := NewHandler(&comp, &repo)

			rr := httptest.NewRecorder()
			req := newCreateOrderRequestWithPayload(t, fmt.Sprintf(`{"size": %d}`, size))

			handler.ServeHTTP(rr, req)

			assertBadRequestResponse(t, rr, "invalid_order_size", "Order sizes must be greater than zero.")
		})
	}
}

func TestServeHTTP_HandleCreateOrder_InternalServerError(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestErrRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateOrderRequestWithPayload(t, `{"size": 1}`)

	handler.ServeHTTP(rr, req)

	assertInternalServerErrorResponse(t, rr)
}

func TestServeHTTP_HandleCreateOrder_Headers(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateOrderRequestWithPayload(t, `{"size": 1}`)

	handler.ServeHTTP(rr, req)

	assertHeader(t, rr, "Content-Type", "application/json")
}

func TestServeHTTP_HandleCreateOrder_CorsHeaders(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodOptions, Path, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	assertStatusOk(t, rr)
	assertCorsHeaders(t, rr)
}

func TestServeHTTP_HandleSetConfig_Success(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateConfigRequestWithPayload(t, `{"pack_sizes": [100, 200]}`)

	handler.ServeHTTP(rr, req)

	assertStatusOk(t, rr)

	cfg := repository.Config{
		PackSizes: []int{100, 200},
	}
	assertRepositoryReceivedConfig(t, repo, cfg)
}

func TestServeHTTP_HandleSetConfig_InvalidPayload(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateConfigRequestWithPayload(t, "{pack_sizes}")

	handler.ServeHTTP(rr, req)

	assertBadRequestResponse(t, rr, "invalid_payload", "Invalid payload.")
}

func TestServeHTTP_HandleSetConfig_InvalidPackSizes(t *testing.T) {
	payloads := []string{"{}", `{"pack_sizes": []}`, `{"pack_sizes": [-1, 100]}`, `{"pack_sizes": [0, 100]}`}

	for _, payload := range payloads {
		t.Run(fmt.Sprintf("with payload: '%s'", payload), func(t *testing.T) {
			comp := TestPackComputer{}
			repo := TestSuccessRepository{}
			handler := NewHandler(&comp, &repo)

			rr := httptest.NewRecorder()
			req := newCreateConfigRequestWithPayload(t, payload)

			handler.ServeHTTP(rr, req)

			assertBadRequestResponse(t, rr, "invalid_pack_sizes", "Pack sizes should have at least one size and all sizes should be greater than zero.")
		})
	}
}

func TestServeHTTP_HandleSetConfig_RemoveDuplicates(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateConfigRequestWithPayload(t, `{"pack_sizes": [100, 100, 200, 200]}`)

	handler.ServeHTTP(rr, req)

	cfg := repository.Config{
		PackSizes: []int{100, 200},
	}
	assertRepositoryReceivedConfig(t, repo, cfg)
}

func TestServeHTTP_HandleSetConfig_InternalServerError(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestErrRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateConfigRequestWithPayload(t, `{"pack_sizes": [100, 200]}`)

	handler.ServeHTTP(rr, req)

	assertInternalServerErrorResponse(t, rr)
}

func TestServeHTTP_HandleSetConfig_Headers(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req := newCreateConfigRequestWithPayload(t, `{"pack_sizes": [100, 200]}`)

	handler.ServeHTTP(rr, req)

	assertHeader(t, rr, "Content-Type", "application/json")
}

func TestServeHTTP_HandleSetConfig_CorsHeaders(t *testing.T) {
	comp := TestPackComputer{}
	repo := TestSuccessRepository{}
	handler := NewHandler(&comp, &repo)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodOptions, Path+configPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	assertStatusOk(t, rr)
	assertCorsHeaders(t, rr)
}

func newCreateOrderRequestWithPayload(t *testing.T, payload string) *http.Request {
	req, err := http.NewRequest(http.MethodPost, Path, bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func newCreateConfigRequestWithPayload(t *testing.T, payload string) *http.Request {
	req, err := http.NewRequest(http.MethodPut, Path+configPath, bytes.NewReader([]byte(payload)))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func assertPackComputerReceivedPackSizes(t *testing.T, comp TestPackComputer, expected []int) {
	slices.Sort(comp.passedPackSizes)
	slices.Sort(expected)
	if !slices.Equal(comp.passedPackSizes, expected) {
		t.Errorf("unexpected pack sizes passed to pack computer: got '%+v' want '%+v'", comp.passedPackSizes, expected)
	}
}

func assertPackComputerReceivedOrderSize(t *testing.T, comp TestPackComputer, expected int) {
	if comp.passedOrderSize != expected {
		t.Errorf("unexpected order size passed to pack computer: got '%+v' want '%+v'", comp.passedOrderSize, expected)
	}
}

func assertOrderPacks(t *testing.T, rr *httptest.ResponseRecorder, expected []pack.Pack) {
	var orderResp Order
	if err := json.NewDecoder(rr.Body).Decode(&orderResp); err != nil {
		t.Fatal(err)
	}

	if ok := pack.EqualSlice(orderResp.Packs, expected); !ok {
		t.Errorf("unexpected packs: got '%+v' want '%+v'", orderResp.Packs, expected)
	}
}

func assertRepositoryReceivedConfig(t *testing.T, repo TestSuccessRepository, expected repository.Config) {
	slices.Sort(repo.passedCfg.PackSizes)
	slices.Sort(expected.PackSizes)
	if !slices.Equal(repo.passedCfg.PackSizes, expected.PackSizes) {
		t.Errorf("unexpected config pack sizes: got '%+v' want '%+v'", repo.passedCfg.PackSizes, expected.PackSizes)
	}
}

func assertStatusOk(t *testing.T, rr *httptest.ResponseRecorder) {
	if rr.Code != http.StatusOK {
		t.Errorf("unexpected status code: got '%d' want '%d'", rr.Code, http.StatusOK)
	}
}

func assertBadRequestResponse(t *testing.T, rr *httptest.ResponseRecorder, expectedErrCode, expectedErrMsg string) {
	if rr.Code != http.StatusBadRequest {
		t.Errorf("unexpected status code: got '%d' want '%d'", rr.Code, http.StatusBadRequest)
	}

	expected := fmt.Sprintf(`{"error_code":"%s","error_message":"%s"}`, expectedErrCode, expectedErrMsg)
	if rr.Body.String() != expected {
		t.Errorf("unexpected body: got '%s' want '%s'", rr.Body.String(), expected)
	}
}

func assertInternalServerErrorResponse(t *testing.T, rr *httptest.ResponseRecorder) {
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: got '%d' want '%d'", rr.Code, http.StatusInternalServerError)
	}

	expected := `{"error_code":"internal_server_error","error_message":"Internal server error."}`
	if rr.Body.String() != expected {
		t.Errorf("unexpected body: got '%s' want '%s'", rr.Body.String(), expected)
	}
}

func assertCorsHeaders(t *testing.T, rr *httptest.ResponseRecorder) {
	assertHeader(t, rr, "Access-Control-Allow-Origin", "*")
	assertHeader(t, rr, "Access-Control-Allow-Methods", "POST, PUT")
	assertHeader(t, rr, "Access-Control-Allow-Headers", "*")
}

func assertHeader(t *testing.T, rr *httptest.ResponseRecorder, name, expected string) {
	headerValue := rr.Header().Get(name)
	if headerValue != expected {
		t.Errorf("unexpected http header: got '%s' want '%s'", headerValue, expected)
	}
}
