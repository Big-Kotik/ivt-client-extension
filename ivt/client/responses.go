package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"inv-client-extension/ivt/types"
	"net/http"
	"os"
)

const (
	tempFilePattern = "save-*.json"
)

func (user *User) SendResponses(responses *types.ResponsesWrapper) {
	for _, res := range responses.Data {
		user.idToRequest.Range(func(key, value interface{}) bool {
			return true
		})

		val, ok := user.idToRequest.LoadAndDelete(res.ID.String())
		if !ok {
			continue
		}
		req, ok := val.(types.Requests)
		if !ok {
			continue
		}

		h := http.Header{}
		for k, vals := range res.Headers {
			for _, val := range vals {
				h.Add(k, val)
			}
		}
		req.WriteHeaders(h)
		req.Write(res.Body)
		req.Complete()
	}
}

func onNewMessageFunc(
	client *telegram.Client,
	user *User,
	logger *zap.Logger,
) func(context.Context, tg.Entities, *tg.UpdateNewMessage) error {
	return func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			fmt.Print("not ok msg")
			return errors.New("not ok msg")
		}

		peer := msg.PeerID
		if !ok {
			fmt.Print("not ok from")
			return errors.New("not ok from")
		}

		uid, ok := peer.(*tg.PeerUser)
		if !ok {
			fmt.Print("not ok uid")
			return errors.New("not ok uid")
		}

		us, _ := client.Self(ctx)
		if uid.UserID == us.ID || msg.FromID != nil {
			return errors.New("it's me")
		}

		media, ok := msg.Media.(*tg.MessageMediaDocument)
		if !ok {
			return errors.New("not ok media")
		}
		docTemp, ok := media.GetDocument()
		if !ok {
			logger.Sugar().Error("Error: not ok")
			return errors.New("not ok media.GetDocument()")
		}
		doc, ok := docTemp.(*tg.Document)
		if !ok {
			logger.Sugar().Error("Error: not ok")
			return errors.New("not ok doc")
		}

		file, err := os.CreateTemp("", tempFilePattern)
		if err != nil {
			return err
		}
		filename := file.Name()
		err = file.Close()
		if err != nil {
			return err
		}

		_, err = downloader.NewDownloader().Download(client.API(), doc.AsInputDocumentFileLocation()).ToPath(ctx, filename)
		if err != nil {
			logger.Sugar().Errorf("Error: %w", err)
			return err
		}
		file, err = os.Open(filename)
		if err != nil {
			logger.Sugar().Errorf("Error: %w", err)
			return err
		}

		var responses types.ResponsesWrapper
		err = json.NewDecoder(file).Decode(&responses)
		if err != nil {
			logger.Sugar().Errorf("Error: %w", err)
			return err
		}
		logger.Sugar().Debugf("Response: %v", responses)

		user.SendResponses(&responses)
		fmt.Println("got message!")
		return nil
	}
}
