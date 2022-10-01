package client

import (
	"context"
	"encoding/json"
	"fmt"
	"inv-client-extension/ivt/types"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
)

func NewRequestWrapper(url string, method string, body []byte, headers map[string][]string) *types.RequestWrapper {
	if body == nil {
		body = make([]byte, 0)
	}
	if headers == nil {
		headers = make(map[string][]string)
	}
	return &types.RequestWrapper{Url: url, Method: method, Headers: headers}
}

func NewRequestsWrapper(requests ...*types.RequestWrapper) types.RequestsWrapper {
	return types.RequestsWrapper{Data: requests}
}

func (user *User) SendRequests(requests ...*types.RequestWrapper) error {
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

func (user *User) Run(c <-chan types.Requests) {
	t := time.NewTicker(1*time.Second + 500*time.Millisecond)
	for range t.C {
		log.Print("HERE")
		requests := make([]types.Requests, 0)
		for {
			flag := false
			select {
			case req := <-c:
				requests = append(requests, req)
			default:
				flag = true
			}
			if flag {
				break
			}
		}

		wrappedRequests := make([]*types.RequestWrapper, 0)
		for _, val := range requests {
			user.idToRequest.Store(val.GetUuid().String(), val)
			log.Printf("store: %s", val.GetUuid().String())
			wrappedRequests = append(wrappedRequests, val.ToRequestWrapper())
		}
		if len(wrappedRequests) == 0 {
			continue
		}
		err := user.SendRequests(wrappedRequests...)
		if err != nil {
			log.Print("err ", err)
			continue
		}
	}
}

func initFile(filename string, requests types.RequestsWrapper) (*os.File, error) {
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
