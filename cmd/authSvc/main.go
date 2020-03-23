package main

import (
	"flag"
	"fmt"
	"github.com/burhon94/alifMux/pkg/mux"
	"github.com/burhon94/authentificationservice/cmd/authSvc/app"
	"github.com/burhon94/authentificationservice/core/auth"
	"github.com/burhon94/bdi/pkg/di"
	jwt "github.com/burhon94/jwt/pkg/core"
	"log"
	"net"
	"net/http"
)

// -authUrl http://localhost:9999 -host 0.0.0.0 -port 10000 -key alifkey

var (
	authUrl = flag.String("authUrl", "", "Auth Service URL")
	host = flag.String("host", "", "Server host")
	port = flag.String("port", "", "Server port")
	secret = flag.String("key", "", "key")
)

func main() {
	flag.Parse()
	addr := net.JoinHostPort(*host, *port)
	keySecret := jwt.Secret(*secret)
	start(addr, keySecret, auth.Url(*authUrl))
}

func start(addr string, secret jwt.Secret, authUrl auth.Url) {
	container := di.NewContainer()
	err := container.Provide(
		app.NewServer,
		mux.NewExactMux,
		func() jwt.Secret { return secret },
		func() auth.Url { return authUrl },
		auth.NewClient,
	)
	if err != nil {
		panic(fmt.Sprintf("can't set provide: %v", err))
	}

	container.Start()
	var appServer *app.Server
	container.Component(&appServer)
	log.Printf("authSvc listinig: ... %s", addr)
	panic(http.ListenAndServe(addr, appServer))
}
