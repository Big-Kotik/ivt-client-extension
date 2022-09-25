package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"inv-client-extension/ivt/types"
	"os"
	"strings"

	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/updates"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

const (
	tempFileName = "save.json"
)

// noSignUp can be embedded to prevent signing up.
type noSignUp struct{}

func (c noSignUp) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("not implemented")
}

func (c noSignUp) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

// termAuth implements authentication via terminal.
type termAuth struct {
	noSignUp

	phone string
}

func (a termAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a termAuth) Password(_ context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func getNonSecretData(msg string) (string, error) {
	fmt.Print(msg)
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (a termAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	return getNonSecretData("Enter code: ")
}

func getPhone() (string, error) {
	return getNonSecretData("Enter phone: ")
}

func (user *User) authClient(logger *zap.Logger) (*telegram.Client, error) {
	//phone, err := getPhone()
	//if err != nil {
	//	return nil, err
	//}
	// TODO: !!!!!
	phone := "+79312741632"

	flow := auth.NewFlow(
		termAuth{phone: phone},
		auth.SendCodeOptions{},
	)

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		Logger:  logger.Named("gaps"),
	})
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger:        logger,
		UpdateHandler: gaps,
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
		},
	})
	if err != nil {
		return client, err
	}

	d.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
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
		// from := msg.FromID.(*tg.PeerUser)

		us, _ := client.Self(ctx)

		if uid.UserID == us.ID || msg.FromID != nil {
			return errors.New("it's me")
		}
		//user, err := client.Self(context.Background())
		//if err != nil {
		//	return err
		//}
		//
		//fullUser, err := client.API().UsersGetFullUser(context.Background(), &tg.InputUser{
		//	UserID:     uid.UserID,
		//	AccessHash: user.AccessHash,
		//})
		//if err != nil {
		//	return err
		//}
		//
		//fullUser.FullUser.

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
		_, err := downloader.NewDownloader().Download(client.API(), doc.AsInputDocumentFileLocation()).ToPath(ctx, tempFileName)
		if err != nil {
			logger.Sugar().Errorf("Error: %w", err)
			return err
		}
		file, err := os.Open(tempFileName)
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

		// log.Printf("%v", responses.Data[0].Body)

		user.SendResponses(&responses)
		fmt.Println("got message!")
		return nil
	})

	_, err = bg.Connect(client)
	if err != nil {
		panic(err)
	}

	if err := client.Auth().IfNecessary(context.Background(), flow); err != nil {
		return client, err
	}

	return client, err
}
