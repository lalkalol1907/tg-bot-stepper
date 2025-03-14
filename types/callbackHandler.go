package types

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CallbackHandler = func(ctx context.Context, b *bot.Bot, update *models.CallbackQuery) error
