package stepper

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/lalkalol1907/tg-bot-stepper/types"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"strings"
)

type Stepper struct {
	cache           types.Cache
	features        map[string]*Feature
	callbackHandler types.CallbackHandler

	singleStepCommands map[string]types.SingleStepCommandHandler

	commandToFeature map[string]string
	logger           *otelzap.Logger
}

func (s *Stepper) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil && s.callbackHandler != nil {
		err := s.callbackHandler(ctx, b, update.CallbackQuery)
		if err != nil {
			s.logger.Ctx(ctx).Warn(fmt.Sprintf("error executing cb command %s for %d: %s", update.CallbackQuery.Data, err.Error()))
		}

		return
	}

	// TODO: check message not null

	chatId := update.Message.From.ID
	text := update.Message.Text

	if strings.HasPrefix(text, "/") {
		ssCommand, ok := s.singleStepCommands[text]
		if ok {
			err := ssCommand(ctx, b, update)
			if err != nil {
				s.logger.Ctx(ctx).Warn(fmt.Sprintf("error executing command %s for %d: %s", text, chatId, err.Error()))
			}
			err = s.cache.Del(ctx, chatId)
			if err != nil {
				s.logger.Ctx(ctx).Warn(fmt.Sprintf("error deleting cache for %d: %s", chatId, err.Error()))
			}
			return
		}
	}

	feature, step, err := s.cache.Get(ctx, chatId)
	if err != nil {
		s.logger.Ctx(ctx).Warn(fmt.Sprintf("error getting cache for %d: %s", chatId, err.Error()))
		return
	}

	// TODO: CallbackQueryHandler

	if feature == nil || strings.HasPrefix(text, "/") {
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
		err = s.cache.Del(ctx, chatId)
		if err != nil {
			s.logger.Ctx(ctx).Warn(fmt.Sprintf("error deleting cache for %d: %s", chatId, err.Error()))
		}
		return
	}

	err = s.cache.Set(ctx, chatId, *feature, *response.NextStep)
	if err != nil {
		s.logger.Ctx(ctx).Warn(fmt.Sprintf("error setting cache for %d: %s", chatId, err.Error()))
	}
}

func (s *Stepper) AddFeature(featureName string, command string, feature *Feature) *Stepper {
	s.features[featureName] = feature
	s.commandToFeature[command] = featureName

	return s
}

// AddInternalFeature Добавляет фичу, которая не имеет команды и омжет триггериться только изнутри
func (s *Stepper) AddInternalFeature(featureName string, feature *Feature) *Stepper {
	s.features[featureName] = feature
	return s
}

func (s *Stepper) AddSingleStepCommand(command string, handler types.SingleStepCommandHandler) *Stepper {
	s.singleStepCommands[command] = handler
	return s
}

func (s *Stepper) AddCallbackHandler(handler types.CallbackHandler) *Stepper {
	s.callbackHandler = handler
	return s
}

func (s *Stepper) OverrideCurrentFeature(ctx context.Context, chatId int64, feature string, nextStep string) error {
	return s.cache.Set(ctx, chatId, feature, nextStep)
}

func NewStepper(cache types.Cache, logger *otelzap.Logger) *Stepper {
	return &Stepper{
		logger: logger,
		cache:  cache,

		features:           make(map[string]*Feature),
		commandToFeature:   make(map[string]string),
		singleStepCommands: make(map[string]types.SingleStepCommandHandler),
	}
}
