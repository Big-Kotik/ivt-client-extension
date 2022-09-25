package client

import (
	"fmt"
	"github.com/gotd/td/telegram"
	"go.uber.org/zap"
	"sync"
)

type User struct {
	botUsername string
	tgClient    *telegram.Client
	idToRequest sync.Map
}

func NewUser(botUsername string, log *zap.Logger) (*User, error) {
	user := &User{botUsername: botUsername}
	client, err := user.authClient(log)
	if err != nil {
		return nil, fmt.Errorf("can't authenticate: %w", err)
	}
	user.tgClient = client
	return user, nil
}
