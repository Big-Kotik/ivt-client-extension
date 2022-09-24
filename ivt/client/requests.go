package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type RequestWrapper struct {
	Url     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

type RequestsWrapper struct {
	Data []RequestWrapper `json:"data"`
}

func NewRequestWrapper(url string, method string, headers map[string]string) RequestWrapper {
	if headers == nil {
		headers = make(map[string]string)
	}
	return RequestWrapper{Url: url, Method: method, Headers: headers}
}

func NewRequestsWrapper(requests ...RequestWrapper) RequestsWrapper {
	return RequestsWrapper{Data: requests}
}

func (user *User) SendRequests(requests ...RequestWrapper) error {
	file, err := initFile("test.json", NewRequestsWrapper(requests...))
	if err != nil {
		return err
	}

	api := tg.NewClient(user.tgClient)
	// Helper for uploading. Automatically uses big file upload when needed.
	u := uploader.NewUploader(api)
	// Helper for sending messages.
	sender := message.NewSender(api).WithUploader(u)
	upload, err := u.FromPath(context.Background(), file.Name())
	if err != nil {
		return fmt.Errorf("upload %q: %w", file.Name(), err)
	}
	// Now we have uploaded file handle, sending it as styled message.
	// First, preparing message.
	document := message.UploadedDocument(upload)
	document.Filename(file.Name())
	// Resolving target. Can be telephone number or @nickname of user,
	// group or channel.
	target := sender.Resolve("@MishaRout")
	if _, err := target.Media(context.Background(), document); err != nil {
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
