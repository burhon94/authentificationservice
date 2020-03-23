package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/burhon94/authentificationservice/core/auth"
	"github.com/burhon94/authentificationservice/core/utils"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func (s *Server) handlerRespHealth() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		_, err := fmt.Fprintf(writer, "Health ok")
		if err != nil {
			log.Printf("err: %v", err)
		}
	}
}

func (s *Server) handleVerifyPage() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "index.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		err := tpl.Execute(writer, struct{}{})
		if err != nil {
			log.Printf("error while executing template %s %v", tpl.Name(), err)
		}
	}
}

func (s *Server) handleUserPage() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates/users", "user.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		err := tpl.Execute(writer, struct{}{})
		if err != nil {
			log.Printf("error while executing template %s %v", tpl.Name(), err)
		}
	}
}

func (s *Server) handleLoginPage() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "login.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		err := tpl.Execute(writer, struct{}{})
		if err != nil {
			log.Printf("error while executing template %s %v", tpl.Name(), err)
		}
	}
}

func (s *Server) handleLogin() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "login.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
		if err != nil {
			log.Printf("error while parse login form: %v", err)
			return
		}

		login := request.PostFormValue("login")
		if login == "" {
			log.Print("login can't be empty")
			return
		}
		password := request.PostFormValue("password")
		if password == "" {
			log.Print("password can't be empty")
			return
		}

		ctx, _ := context.WithTimeout(request.Context(), time.Second)
		token, err := s.authClient.Login(ctx, login, password)
		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				log.Print("auth service didn't response in given time")
				log.Print("another err")
			case errors.Is(err, context.Canceled):
				log.Print("auth service didn't response in given time")
				log.Print("another err")
			case errors.Is(err, auth.ErrResponse):
				var typedErr *auth.ErrorResponse
				ok := errors.As(err, &typedErr)
				if ok {
					tplData := struct {
						Err string
					}{
						Err: "",
					}

					if utils.StringInSlice("err.password_mismatch", typedErr.Errors) {
						tplData.Err = "err.password_mismatch"
					}

					err := tpl.Execute(writer, tplData)
					if err != nil {
						log.Print(err)
					}
				}
			}
			return
		}
		if token != "" {
		cookie := &http.Cookie{
			Name:     "token",
			Value:    token,
			HttpOnly: true,
		}
		http.SetCookie(writer, cookie)

	}
		http.Redirect(writer, request, Root, http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		cookie := &http.Cookie{
			Name:     "token",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		}
		http.SetCookie(writer, cookie)

		http.Redirect(writer, request, Root, http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleRegisterPage()http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "register.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request){
		err := tpl.Execute(writer, nil)
		if err != nil {
			log.Print(err)
		}
	}
}

func (s *Server) handleRegister() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err:= request.ParseForm()
		if err != nil {
			log.Print(err)
		}

		name := request.PostForm.Get("name")
		login := request.PostForm.Get("login")
		password := request.PostForm.Get("password")
		if name == "" {
			return
		}

		if login == "" {
			return
		}

		if password == "" {
			return
		}


		ctx, _ := context.WithTimeout(request.Context(), time.Second)
		err = s.authClient.Register(ctx, name, login, password)
		if err != nil {
			if err == auth.ErrAddNewUser {
				_, _ = writer.Write([]byte("this user is exist, use any"))
				return
			} else {
				log.Printf("Error, %v", err)
			}
		}else{
			http.Redirect(writer, request, Root, http.StatusTemporaryRedirect)
		}
	}
}

func (s *Server) handlePageAfterAuth() http.HandlerFunc {
	var (
		tpl *template.Template
		err error
	)
	tpl, err = template.ParseFiles(filepath.Join("web/templates", "register.gohtml"))
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		err := tpl.Execute(writer, nil)
		if err != nil {
			log.Print(err)
		}
	}
}

