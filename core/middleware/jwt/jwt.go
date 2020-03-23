package jwt

import (
	"context"
	jwtcore "github.com/burhon94/jsonwebtoken/pkg/cmd"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type contextKey string // int?
var payloadContextKey = contextKey("jwt")

const (
	SourceAuthorization = iota
	SourceCookie
)

func JWT(source int, redirect bool, redirectUrl string, payloadType reflect.Type, secret jwtcore.Secret) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			token := ""

			switch source {
			case SourceAuthorization:
				header := request.Header.Get("Authorization")
				if header == "" {
					break
				}
				if !strings.HasPrefix(header, "Bearer ") {
					break
				}
				token = header[len("Bearer "):]
			case SourceCookie:
				cookie, err := request.Cookie("token")
				if err != nil {
					if err == http.ErrNoCookie {
						break
					}
					break
				}
				token = cookie.Value
			}

			if token == "" {
				next(writer, request)
				return
			}

			ok, err := jwtcore.Verify(token, secret)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if !ok {
				if redirect {
					http.Redirect(writer, request, redirectUrl, http.StatusTemporaryRedirect)
					return
				}
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			payload := reflect.New(payloadType).Interface()

			err = jwtcore.Decode(token, payload)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			ok, err = jwtcore.IsNotExpired(payload, time.Now())
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if !ok {
				http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			log.Print(payload)

			ctx := context.WithValue(request.Context(), payloadContextKey, payload)
			next(writer, request.WithContext(ctx))
		}
	}
}

func FromContext(ctx context.Context) (payload interface{}) {
	payload = ctx.Value(payloadContextKey)
	return
}

func IsContextNonEmpty(ctx context.Context) bool {
	return nil != ctx.Value(payloadContextKey)
}
