package metcoll

import (
	"context"
	"net/http"

	"github.com/ArtemShalinFe/metcoll/internal/configuration"
)

type Server struct {
	*http.Server
}

func (s *Server) Interrupt() error {

	// пакет context же пока не проходили - пока todo поиспользую
	if err := s.Shutdown(context.TODO()); err != nil {
		return err
	}

	return nil

}

func NewServer(cfg *configuration.Config) *Server {
	s := http.Server{
		Addr: cfg.Address,
	}
	return &Server{&s}
}
