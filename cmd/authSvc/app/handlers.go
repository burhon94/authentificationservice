package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/burhon94/alifMux/pkg/mux"
	"github.com/burhon94/authentificationservice/core/auth"
	"github.com/burhon94/authentificationservice/core/utils"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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

		userData, err := s.authClient.GetUserData(request)
		err = tpl.Execute(writer, userData)
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

func (s *Server) handleRegisterPage() http.HandlerFunc {
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

func (s *Server) handleRegister() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
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
		} else {
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

func (s *Server) handleUserEdit() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx, _ := context.WithTimeout(request.Context(), time.Second*2)
		value, ok := mux.FromContext(ctx, "id")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(value)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		nameSurname := request.FormValue("nameSurname")
		if nameSurname == "" {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = s.authClient.UpdateUser(ctx, request, int64(id), nameSurname)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		http.Redirect(writer, request, MePage, http.StatusFound)
	}
}

func (s *Server) handleUserPassEdit() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx, _ := context.WithTimeout(request.Context(), time.Second)
		value, ok := mux.FromContext(ctx, "id")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(value)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		oldPass := request.FormValue("oldPass")
		pass := request.FormValue("pass")
		pass2 := request.FormValue("pass2")
		if pass != pass2 {
			_, err := fmt.Fprintf(writer, "new password: %s not mismatched: %s", pass, pass2)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		err = s.authClient.CheckPass(ctx, int64(id), oldPass, request)
		if err != nil {
			if !errors.Is(err, errors.New("wrong password")) {
				http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		err = s.authClient.UpdatePass(ctx, int64(id), pass, request)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(writer, request, MePage, http.StatusFound)
	}
}

func (s *Server) handleUserAvatarEdit() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseMultipartForm(10 * 1024 * 1024 * 1024)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		ctx, _ := context.WithTimeout(request.Context(), time.Second)

		file, header, err := request.FormFile("image")
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		defer func() {
			if file.Close() != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}()
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		avatarUrl, err := s.fileClient.UploadFile(ctx, bytes, header.Filename)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		value, ok := mux.FromContext(ctx, "id")
		if !ok {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(value)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = s.authClient.UpdateAvatar(ctx, id, avatarUrl, request)
		if err != nil {
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		http.Redirect(writer, request, MePage, http.StatusFound)
	}
}
