package client

import (
	"fmt"
	"time"

	"github.com/gotd/td/telegram"
	"go.uber.org/zap"
)

type User struct {
	botUsername string
	tgClient    *telegram.Client
}

func NewUser(botUsername string, log *zap.Logger) (*User, error) {
	client, err := authClient(log)
	if err != nil {
		return nil, fmt.Errorf("can't authenticate: %w", err)
	}
	return &User{botUsername: botUsername, tgClient: client}, nil
}

func (user *User) Run() {
	t := time.NewTicker(10 * time.Second)
	for range t.C {
		fmt.Println("Ticking...")
	}
}
