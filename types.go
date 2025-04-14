package tg_bot_stepper

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type StepExecutionResult struct {
	IsFinal  bool
	NextStep string
}

type Cache interface {
	Set(ctx context.Context, chatId int64, featureName string, stepName string) error
	Get(ctx context.Context, chatId int64) (string, string, error)
	Del(ctx context.Context, chatId int64) error
}

type SingleStepCommandHandler = func(ctx context.Context, b *bot.Bot, update *models.Update) error
type CallbackHandler = func(ctx context.Context, b *bot.Bot, update *models.CallbackQuery) error
type Step = func(ctx context.Context, b *bot.Bot, update *models.Update) (StepExecutionResult, error)
