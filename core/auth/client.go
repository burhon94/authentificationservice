package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Url string

type ReqForToken struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestNewUsr struct {
	Username string `json:"username"`
	UserLogin string `json:"user_login"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
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

type Client struct {
	url Url
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
		Username:     name,
		UserLogin:    login,
		Password: password,
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


