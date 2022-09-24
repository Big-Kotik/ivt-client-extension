package client

import (
	"fmt"

	"github.com/gotd/td/telegram"
	"go.uber.org/zap"
)

type User struct {
	botId    int
	tgClient *telegram.Client
}

func NewUser(botId int, log *zap.Logger) (*User, error) {
	client, err := authClient(log)
	if err != nil {
		return nil, fmt.Errorf("can't authenticate: %w", err)
	}
	return &User{botId: botId, tgClient: client}, nil
}
