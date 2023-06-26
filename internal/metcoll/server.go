package metcoll

import (
	"net/http"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
)

type Server struct {
	*http.Server
}

func NewServer(cfg *configuration.Config) *Server {
	s := http.Server{
		Addr: cfg.Address,
	}
	return &Server{&s}
}
