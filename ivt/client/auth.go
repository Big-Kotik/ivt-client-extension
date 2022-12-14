package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram/updates"

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

func (user *User) authClient(logger *zap.Logger) (*telegram.Client, error) {
	phone, err := getPhone()
	if err != nil {
		return nil, err
	}

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

	d.OnNewMessage(onNewMessageFunc(client, user, logger))

	_, err = bg.Connect(client)
	if err != nil {
		panic(err)
	}

	if err := client.Auth().IfNecessary(context.Background(), flow); err != nil {
		return client, err
	}

	return client, err
}
