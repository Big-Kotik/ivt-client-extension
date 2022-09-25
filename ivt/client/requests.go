package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
)

type RequestWrapper struct {
	Url     string              `json:"url"`
	Method  string              `json:"method"`
	Body    []byte              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

type RequestsWrapper struct {
	Data []RequestWrapper `json:"data"`
}

type Wrapper interface {
	ToRequestWrapper() *RequestWrapper
}

func NewRequestWrapper(url string, method string, body []byte, headers map[string][]string) RequestWrapper {
	if body == nil {
		body = make([]byte, 0)
	}
	if headers == nil {
		headers = make(map[string][]string)
	}
	return RequestWrapper{Url: url, Method: method, Headers: headers}
}

func NewRequestsWrapper(requests ...RequestWrapper) RequestsWrapper {
	return RequestsWrapper{Data: requests}
}

func (user *User) SendRequests(requests ...RequestWrapper) error {
	file, err := initFile("test.json", NewRequestsWrapper(requests...))
	defer file.Close()
	if err != nil {
		return err
	}

	api := user.tgClient.API()
	u := uploader.NewUploader(api)
	sender := message.NewSender(api).WithUploader(u)
	upload, err := u.FromPath(context.Background(), file.Name())
	if err != nil {
		return fmt.Errorf("upload %q: %w", file.Name(), err)
	}
	target := sender.Resolve(user.botUsername)
	if _, err := target.File(context.Background(), upload); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func initFile(filename string, requests RequestsWrapper) (*os.File, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("can't find cahce dir: %w", err)
	}
	filePath := filepath.Join(cacheDir, filename)
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return nil, fmt.Errorf("can't open or create file %s: %w", filePath, err)
	}

	marshal, err := json.MarshalIndent(requests, "", "")
	if err != nil {
		return nil, err
	}
	err = file.Truncate(0)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	_, err = file.Write(marshal)
	if err != nil {
		return nil, err
	}

	return file, nil
}
