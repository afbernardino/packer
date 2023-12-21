package rest

import (
	"errors"
	"log"
	"net/http"
	"packer/internal/rest/order"
	"packer/internal/rest/order/pack"
	"strconv"
	"time"
)

type Config struct {
	ReadTimeout, WriteTimeout, IdleTimeout time.Duration
}

// ApiService handles incoming HTTP requests and can use an order's repository.
type ApiService struct {
	cfg  Config
	repo order.Repository
}

func NewApiService(cfg Config, repo order.Repository) ApiService {
	return ApiService{
		cfg:  cfg,
		repo: repo,
	}
}

func (svc *ApiService) Serve(port int) {
	s := svc.newServer(port)

	log.Printf("listening at port %d...\n", port)

	if err := s.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("the api service will be shutdown because an error ocurred: %v\n", err)
		}
	}
}

func (svc *ApiService) newServer(port int) http.Server {
	return http.Server{
		Addr:         ":" + strconv.Itoa(port),
		ReadTimeout:  svc.cfg.ReadTimeout,
		WriteTimeout: svc.cfg.WriteTimeout,
		IdleTimeout:  svc.cfg.IdleTimeout,
		Handler:      svc.newServeMux(),
	}
}

func (svc *ApiService) newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	computer := pack.NewComputer()
	orderHandler := order.NewHandler(&computer, svc.repo)
	mux.Handle(order.Path, http.StripPrefix(order.Path, &orderHandler))
	mux.Handle(order.Path+"/", http.StripPrefix(order.Path, &orderHandler))
	return mux
}
