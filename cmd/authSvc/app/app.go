package app

import (
	"github.com/burhon94/alifMux/pkg/mux"
	"github.com/burhon94/authentificationservice/core/auth"
	jwt "github.com/burhon94/jwt/pkg/core"
	"net/http"
)

type Server struct {
	router     *mux.ExactMux
	secret     jwt.Secret
	authClient *auth.Client
}

// dig - check nil
func NewServer(router *mux.ExactMux, secret jwt.Secret, authClient *auth.Client) *Server {
	return &Server{router: router, secret: secret, authClient: authClient}
}

func (s *Server) Start() {
	s.InitRoutes()
}

type ErrorDTO struct {
	Errors []string `json:"errors"`
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}
