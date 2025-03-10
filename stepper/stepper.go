package stepper

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/lalkalol1907/tg-bot-stepper/types"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type Stepper struct {
	cache    types.Cache
	features map[string]*Feature

	commandToFeature map[string]string
	logger           *otelzap.Logger
}

func (s *Stepper) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	userId := update.Message.From.ID
	chatId, err := s.cache.GetChatIdByUserId(ctx, userId)
	if err != nil || chatId == nil {
		newChatId := update.Message.Chat.ID
		err = s.cache.SaveUserIdChatId(ctx, newChatId, userId)
		chatId = &newChatId
	}
	if err != nil {
		s.logger.Ctx(ctx).Warn(fmt.Sprintf("error setting cache for %d: %s", chatId, err.Error()))
		return
	}

	feature, step, err := s.cache.Get(ctx, *chatId)
	if err != nil {
		s.logger.Ctx(ctx).Warn(fmt.Sprintf("error getting cache for %d: %s", chatId, err.Error()))
		return
	}

	if feature == nil {
		text := update.Message.Text
		newFeature, ok := s.commandToFeature[text]

		if !ok {
			s.logger.Ctx(ctx).Warn(fmt.Sprintf("error parsing command for %d, text: %s", chatId, text))
			return
		}

		feature = &newFeature
	}

	response, err := s.features[*feature].Run(ctx, step, b, update)
	if err != nil {
		s.logger.Ctx(ctx).Error(fmt.Sprintf(
			"error running step %s of feature %s for %d: %s",
			*step,
			*feature,
			chatId,
			err.Error(),
		))
		return
	}

	if response.IsFinal {
		err = s.cache.Del(ctx, *chatId)
		if err != nil {
			s.logger.Ctx(ctx).Warn(fmt.Sprintf("error deleting cache for %d: %s", chatId, err.Error()))
		}
		return
	}

	err = s.cache.Set(ctx, *chatId, *feature, *response.NextStep)
	if err != nil {
		s.logger.Ctx(ctx).Warn(fmt.Sprintf("error setting cache for %d: %s", chatId, err.Error()))
	}
}

func (s *Stepper) AddFeature(featureName string, command string, feature *Feature) *Stepper {
	s.features[featureName] = feature
	s.commandToFeature[command] = featureName

	return s
}

func NewStepper(cache types.Cache, logger *otelzap.Logger) *Stepper {
	return &Stepper{
		logger: logger,
		cache:  cache,

		features:         make(map[string]*Feature),
		commandToFeature: make(map[string]string),
	}
}
