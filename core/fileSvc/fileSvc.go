package fileSvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type Url string

type FileClient struct {
	url Url
}

func NewFileClient(url Url) *FileClient {
	return &FileClient{url: url}
}

type fileStruct struct {
	FileName string `json:"fileName"`
}

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

var ErrUnknown = errors.New("unknown error")

func (f *FileClient) UploadFile(ctx context.Context, fileBytes []byte, filename string) (fileUrl string, err error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("files", filename) //give file a name
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = fw.Write(fileBytes)
	if err != nil { //copy the file to the multipart buffer
		fmt.Println(err)
		return
	}

	err = w.Close()
	if err != nil {
		return "", nil
	}

	// Upload file
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/files/", f.url),
		&b,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	defer func() {
		if err := response.Body.Close(); err != nil {
			return
		}
	}()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("can't parse response: %w", err)
	}

	switch response.StatusCode {
	case 200:
		var dataResponse fileStruct
		err := json.Unmarshal(responseBody, &dataResponse)
		if err != nil {
			return"", err
		}
		return dataResponse.FileName, nil
	case 500:
		return "", errors.New("can't decode response")
	default:
		return "", ErrUnknown
	}
}
