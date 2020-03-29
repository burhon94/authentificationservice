package app

import (
	"github.com/burhon94/authentificationservice/core/middleware/authenticated"
	"github.com/burhon94/authentificationservice/core/middleware/jwt"
	"github.com/burhon94/authentificationservice/core/middleware/logger"
	"github.com/burhon94/authentificationservice/core/middleware/unauthenticated"
	jwt2 "github.com/burhon94/jsonwebtoken/pkg/cmd"
	"reflect"
)

var (
	Root   = "http://localhost:4444"
	MePage = "/me"
	Login  = "/login"
	Logout = "/logout"
)

func (s *Server) InitRoutes() {
	jwtMW := jwt.JWT(jwt.SourceCookie, true, Logout, reflect.TypeOf((*Payload)(nil)).Elem(), jwt2.Secret(s.secret))
	authMW := authenticated.Authenticated(jwt.IsContextNonEmpty, true, Root)
	unAuthMW := unauthenticated.Unauthenticated(jwt.IsContextNonEmpty, true, MePage)
	s.router.GET("/api/health", s.handlerRespHealth(), logger.Logger("GET/HEALTH"))

	// registrations
	s.router.POST("/register", s.handleRegister(), unAuthMW, jwtMW, logger.Logger("HTTP"))
	s.router.GET("/register", s.handleRegisterPage(), unAuthMW, jwtMW, logger.Logger("HTTP"))

	// verify - identify
	s.router.GET("/api/verify", s.handleVerifyPage(), unAuthMW, jwtMW, logger.Logger("HTTP"))
	s.router.POST("/api/verify", s.handleUserPage(), authMW, jwtMW, logger.Logger("HTTP"))
	s.router.GET(Login, s.handleLoginPage(), unAuthMW, jwtMW, logger.Logger("HTTP"))
	s.router.POST(Login, s.handleLogin(), unAuthMW, jwtMW, logger.Logger("HTTP"))
	s.router.GET(Logout, s.handleLogout(), authMW, jwtMW, logger.Logger("HTTP"))
	s.router.POST(Logout, s.handleLogout(), authMW, jwtMW, logger.Logger("HTTP"))

	// GET -> UserPage
	s.router.GET("/me", s.handleUserPage(), authMW, jwtMW, logger.Logger("HTTP"))
	s.router.POST("/me", s.handleUserPage(), unAuthMW, jwtMW, logger.Logger("HTTP"))

	s.router.POST("/user/{id}", s.handleUserEdit(), authMW, jwtMW, logger.Logger("EDIT_USER"))
	s.router.POST("/user/{id}/new_pass", s.handleUserPassEdit(), authMW, jwtMW, logger.Logger("USER_PASS_EDIT"))
}
