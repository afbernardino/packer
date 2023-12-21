package order

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"packer/internal/rest/order/pack"
	"packer/internal/rest/order/repository"
	"strings"
)

const (
	Path       = "/orders"
	configPath = "/config"
)

// PacksComputer computes the number of packs in an order.
type PacksComputer interface {
	ComputePacks(packSizes []int, orderSize int) []pack.Pack
}

type Repository interface {
	SetConfig(ctx context.Context, cfg repository.Config) error
	FindConfig(ctx context.Context) (repository.Config, error)
}

type Handler struct {
	packsComputer PacksComputer
	repository    Repository
}

func NewHandler(packsComputer PacksComputer, repository Repository) Handler {
	return Handler{
		packsComputer: packsComputer,
		repository:    repository,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setHeaders(w) // could be more granular

	switch {
	case r.Method == http.MethodPost:
		h.handleCreateOrder(w, r)
	case r.Method == http.MethodPut && isConfigPath(r.URL.Path):
		h.handleSetConfig(w, r)
	}
}

func setHeaders(w http.ResponseWriter) {
	setCorsHeaders(w)
	w.Header().Set("Content-Type", "application/json")
}

func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

func isConfigPath(path string) bool {
	return path[strings.LastIndex(path, "/"):] == configPath
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var orderReq Request
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		h.writeBadRequestResponse(w, errRespInvalidPayload)
		return
	}

	if orderReq.Size <= 0 {
		h.writeBadRequestResponse(w, errRespOrderSize)
		return
	}

	packs, err := h.computePacks(ctx, orderReq.Size)
	if err != nil {
		log.Println(err)
		h.writeInternalServerErrorResponse(w)
		return
	}

	jsonBytes, err := json.Marshal(Order{Packs: packs})
	if err != nil {
		log.Println(err)
		h.writeInternalServerErrorResponse(w)
		return
	}

	h.writeMessageWithStatusResponse(w, jsonBytes, http.StatusOK)
}

func (h *Handler) computePacks(ctx context.Context, orderSize int) ([]pack.Pack, error) {
	cfg, err := h.repository.FindConfig(ctx)
	if err != nil {
		return nil, err
	}

	return h.packsComputer.ComputePacks(cfg.PackSizes, orderSize), nil
}

func (h *Handler) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cfg Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		h.writeBadRequestResponse(w, errRespInvalidPayload)
		return
	}

	if !pack.SizesValid(cfg.PackSizes) {
		h.writeBadRequestResponse(w, errRespInvalidPackSizes)
		return
	}

	if err := h.saveConfig(ctx, cfg); err != nil {
		log.Println(err)
		h.writeInternalServerErrorResponse(w)
		return
	}

	jsonBytes, err := json.Marshal(cfg)
	if err != nil {
		log.Println(err)
		h.writeInternalServerErrorResponse(w)
		return
	}

	h.writeMessageWithStatusResponse(w, jsonBytes, http.StatusOK)
}

func (h *Handler) saveConfig(ctx context.Context, cfg Config) error {
	packSizes := pack.RemoveDuplicateSizes(cfg.PackSizes)
	return h.repository.SetConfig(ctx, repository.Config{PackSizes: packSizes})
}

func (h *Handler) writeBadRequestResponse(w http.ResponseWriter, errResp ErrorResponse) {
	jsonBytes, err := json.Marshal(errResp)
	if err != nil {
		log.Println(err)
		h.writeInternalServerErrorResponse(w)
		return
	}
	h.writeMessageWithStatusResponse(w, jsonBytes, http.StatusBadRequest)
}

func (h *Handler) writeInternalServerErrorResponse(w http.ResponseWriter) {
	jsonBytes, err := json.Marshal(errRespInternalServerError)
	if err != nil {
		log.Println(err)
		return
	}
	h.writeMessageWithStatusResponse(w, jsonBytes, http.StatusInternalServerError)
}

func (h *Handler) writeMessageWithStatusResponse(w http.ResponseWriter, message []byte, status int) {
	w.WriteHeader(status)
	_, err := w.Write(message)
	if err != nil {
		log.Println(err)
	}
}
