package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/updates"
	"os"
	"strings"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
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

func authClient(log *zap.Logger) (*telegram.Client, error) {
	//phone, err := getPhone()
	//if err != nil {
	//	return nil, err
	//}
	// TODO: !!!!!
	phone := ""

	flow := auth.NewFlow(
		termAuth{phone: phone},
		auth.SendCodeOptions{},
	)

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		Logger:  log.Named("gaps"),
	})
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger:        log,
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
			return errors.New("")
		}

		peer := msg.PeerID
		if !ok {
			fmt.Print("not ok from")
			return errors.New("")
		}

		uid, ok := peer.(*tg.PeerUser)
		if !ok {
			fmt.Print("not ok uid")
			return errors.New("")
		}
		if uid.UserID == 5365342933 {
			fmt.Println("It's me")
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
			return errors.New("")
		}
		docTemp, ok := media.GetDocument()
		if !ok {
			return errors.New("")
		}

		doc, ok := docTemp.(*tg.Document)
		if !ok {
			return errors.New("")
		}
		_, err := downloader.NewDownloader().Download(client.API(), doc.AsInputDocumentFileLocation()).ToPath(ctx, "save.json")
		if err != nil {
			return err
		}

		fmt.Print("got message!" + string(doc.FileReference))
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
