package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/burhon94/jwt"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Url string

type Client struct {
	url Url
}

type ReqForToken struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestNewUsr struct {
	Username  string `json:"username"`
	UserLogin string `json:"user_login"`
	Password  string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type UserStruct struct {
	Id  int64 `json:"id"`
	Exp int64 `json:"exp"`
}

type userId struct {
	Id int64 `json:"id"`
}

type userIdName struct {
	Id          int64  `json:"id"`
	NameSurname string `json:"name_surname"`
}

type userIdPass struct {
	Id   int64  `json:"id"`
	Pass string `json:"pass"`
}

type userIdAvatar struct {
	Id   int64  `json:"id"`
	AvatarUrl string `json:"avatar_url"`
}

type ResponseDTO struct {
	Id          int64    `json:"id"`
	Login       string   `json:"login"`
	NameSurname string   `json:"name_surname"`
	Avatar      string   `json:"avatar"`
	Role        []string `json:"role"`
}

type ResponseChangeDTO struct {
	Id          int64  `json:"id"`
	Password    string `json:"password"`
	NameSurname string `json:"name_surname"`
	Avatar      string `json:"avatar"`
}

var ErrUnknown = errors.New("unknown error")
var ErrResponse = errors.New("response error")
var ErrAddNewUser = errors.New("allready is exist user by this login")

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

func (e *ErrorResponse) Error() string {
	return strings.Join(e.Errors, ", ")
}

// for errors.Is
func (e *ErrorResponse) Unwrap() error {
	return ErrResponse
}

func NewClient(url Url) *Client {
	return &Client{url: url}
}

func (c *Client) Login(ctx context.Context, login string, password string) (token string, err error) {
	requestData := ReqForToken{
		Username: login,
		Password: password,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("can't encode requestBody %v: %w", requestData, err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/tokens", c.url),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return "", fmt.Errorf("can't create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("can't send request: %w", err)
	}
	defer func() {
		err = response.Body.Close()
		if err != nil {
			return
		}
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("can't parse response: %w", err)
	}

	switch response.StatusCode {
	case 200:
		var responseData *TokenResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return "", fmt.Errorf("can't decode response: %w", err)
		}
		return responseData.Token, nil
	case 400:
		var responseData *ErrorResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return "", fmt.Errorf("can't decode response: %w", err)
		}
		return "", responseData
	default:
		return "", ErrUnknown
	}

}

func (c *Client) Register(ctx context.Context, name, login, password string) (err error) {
	requestData := RequestNewUsr{
		Username:  name,
		UserLogin: login,
		Password:  password,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("can't encode requestBody %v: %w", requestData, err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/users/0", c.url),
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return fmt.Errorf("can't create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("can't send request: %w", err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			return
		}
	}()

	switch response.StatusCode {
	case 200:
		return nil
	case 400:
		return ErrAddNewUser
	default:
		return ErrUnknown
	}

}

func (c *Client) GetUserData(request *http.Request) (userData ResponseDTO, err error) {
	cookie, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return userData, err
		}
		return userData, err
	}

	var userDataToken *UserStruct
	err = jwt.Decode(cookie.Value, &userDataToken)

	id := userDataToken.Id

	bodyReq, err := json.Marshal(&userId{id})
	if err != nil {
		return userData, err
	}

	reqCtx, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("%s/api/users/me", c.url),
		bytes.NewBuffer(bodyReq),
	)
	if err != nil {
		return userData, err
	}
	reqCtx.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cookie.Value))
	reqCtx.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(reqCtx)
	if err != nil {
		return userData, err
	}
	defer func() {
		if err = response.Body.Close(); err != nil {
			return
		}
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return userData, nil
	}

	switch response.StatusCode {
	case 200:
		err = json.Unmarshal(responseBody, &userData)
		if err != nil {
			return userData, fmt.Errorf("can't decode response: %w", err)
		}

		log.Print(userData)

		return userData, nil
	case 400:
		var responseData *ErrorResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return userData, fmt.Errorf("can't decode response: %w", err)
		}
		return userData, responseData
	default:
		return userData, ErrUnknown
	}
}

func (c *Client) UpdateUser(ctx context.Context, formRequest *http.Request, idUser int64, formNameSurname string) (err error) {
	cookie, err := formRequest.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return err
		}
		return err
	}

	var reqUpdateUserData userIdName
	reqUpdateUserData.Id = idUser
	reqUpdateUserData.NameSurname = formNameSurname
	bodyReq, err := json.Marshal(reqUpdateUserData)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/users/%d/edit", c.url, idUser),
		bytes.NewBuffer(bodyReq),
	)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cookie.Value))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			return
		}
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case 200:
		return nil
	//err := bcrypt.CompareHashAndPassword([]byte(dataUser.Password), []byte(formOldpass))
	//if err != nil {
	//	return errors.New("wrong password")
	//}
	//
	//err = bcrypt.CompareHashAndPassword([]byte(dataUser.Password), []byte(formNewPass))
	//if err == nil {
	//	//TODO UPDATE PASSWORD ON DB
	//}
	//
	//if dataUser.NameSurname != formNameSurname {
	//	//TODO UPDATE USER NAME AND SURNAME ON DB
	//}
	//
	//if file.Filename != "" {
	//	//TODO UPLOAD FILE TO FILE SVC
	//	//TODO UPDATE USER AVATAR ON DB
	//}
	case 400:
		var responseData *ErrorResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return fmt.Errorf("can't decode response: %w", err)
		}
		return responseData
	default:
		return ErrUnknown
	}
}

func (c *Client) CheckPass(ctx context.Context, id int64, oldPass string, formRequest *http.Request) (err error) {
	cookie, err := formRequest.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return err
		}
		return err
	}

	var reqCheckPass userId
	reqCheckPass.Id = id
	bodyReq, err := json.Marshal(reqCheckPass)
	if err != nil {
		return err
	}

	requestCheckerPass, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/users/%d/pass", c.url, id),
		bytes.NewBuffer(bodyReq),
	)
	if err != nil {
		return err
	}

	requestCheckerPass.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cookie.Value))
	requestCheckerPass.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(requestCheckerPass)
	if err != nil {
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			return
		}
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case 200:
		var dataResponse userIdPass
		err := json.Unmarshal(responseBody, &dataResponse)
		if err != nil {
			return err
		}

		err = bcrypt.CompareHashAndPassword([]byte(dataResponse.Pass), []byte(oldPass))
		if err != nil {
			return errors.New("wrong password")
		}
		return nil
	case 400:
		var responseData *ErrorResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return fmt.Errorf("can't decode response: %w", err)
		}
		return responseData
	default:
		return ErrUnknown
	}
}

func (c *Client) UpdatePass(ctx context.Context, id int64, NewPass string, formRequest *http.Request) (err error) {
	cookie, err := formRequest.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return err
		}
		return err
	}

	var reqUpdatePass userIdPass
	reqUpdatePass.Id = id
	reqUpdatePass.Pass = NewPass
	bodyReq, err := json.Marshal(reqUpdatePass)
	if err != nil {
		return err
	}

	requestUpdatePass, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/users/%d/edit/pass", c.url, id),
		bytes.NewBuffer(bodyReq),
	)
	if err != nil {
		return err
	}

	requestUpdatePass.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cookie.Value))
	requestUpdatePass.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(requestUpdatePass)
	if err != nil {
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			return
		}
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case 200:
		return nil
	case 400:
		var responseData *ErrorResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return fmt.Errorf("can't decode response: %w", err)
		}
		return responseData
	default:
		return ErrUnknown
	}
}

func (c *Client) UpdateAvatar(ctx context.Context, id int, urlAvatar string, formRequest *http.Request) (err error) {
	cookie, err := formRequest.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			return err
		}
		return err
	}

	var reqUpdatePass userIdAvatar
	reqUpdatePass.Id = int64(id)
	reqUpdatePass.AvatarUrl = urlAvatar
	bodyReq, err := json.Marshal(reqUpdatePass)
	if err != nil {
		return err
	}

	requestUpdatePass, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/users/%d/edit/avatar", c.url, id),
		bytes.NewBuffer(bodyReq),
	)
	if err != nil {
		return err
	}

	requestUpdatePass.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cookie.Value))
	requestUpdatePass.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(requestUpdatePass)
	if err != nil {
		return err
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			return
		}
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case 200:
		return nil
	case 400:
		var responseData *ErrorResponse
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			return fmt.Errorf("can't decode response: %w", err)
		}
		return responseData
	default:
		return ErrUnknown
	}
}
