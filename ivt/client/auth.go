package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// noSignUp can be embedded to prevent signing up.
type noSignUp struct{}

func (c noSignUp) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("not implemented")
}

func (c noSignUp) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
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
	phone, err := getPhone()
	if err != nil {
		return nil, err
	}

	// Setting up authentication flow helper based on terminal auth.
	flow := auth.NewFlow(
		termAuth{phone: phone},
		auth.SendCodeOptions{},
	)

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: log,
	})
	if err != nil {
		return client, err
	}
	err = client.Run(context.Background(), func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		log.Info("Success")

		return nil
	})
	return client, err
}
