package stepper

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	tbs "github.com/lalkalol1907/tg-bot-stepper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"strings"
)

type Stepper struct {
	cache           tbs.Cache
	features        map[string]*Feature
	callbackHandler tbs.CallbackHandler

	singleStepCommands map[string]tbs.SingleStepCommandHandler

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

	if update.Message == nil {
		s.logger.Ctx(ctx).Warn("empty message")
		return
	}

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

	if len(feature) == 0 || strings.HasPrefix(text, "/") {
		newFeature, ok := s.commandToFeature[text]

		if !ok {
			s.logger.Ctx(ctx).Warn(fmt.Sprintf("error parsing command for %d, text: %s", chatId, text))
			return
		}

		feature = newFeature
	}

	f, ok := s.features[feature]
	if !ok {
		otelzap.Ctx(ctx).Error("no feature found", zap.String("feature", feature))
		return
	}

	response, err := f.Run(ctx, step, b, update)
	if err != nil {
		s.logger.Ctx(ctx).Error(fmt.Sprintf(
			"error running step feature %s for %d: %s",
			feature,
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

	err = s.cache.Set(ctx, chatId, feature, response.NextStep)
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

func (s *Stepper) AddSingleStepCommand(command string, handler tbs.SingleStepCommandHandler) *Stepper {
	s.singleStepCommands[command] = handler
	return s
}

func (s *Stepper) AddCallbackHandler(handler tbs.CallbackHandler) *Stepper {
	s.callbackHandler = handler
	return s
}

func (s *Stepper) OverrideCurrentFeature(ctx context.Context, chatId int64, feature string, nextStep string) error {
	return s.cache.Set(ctx, chatId, feature, nextStep)
}

func NewStepper(cache tbs.Cache, logger *otelzap.Logger) *Stepper {
	return &Stepper{
		logger: logger,
		cache:  cache,

		features:           make(map[string]*Feature),
		commandToFeature:   make(map[string]string),
		singleStepCommands: make(map[string]tbs.SingleStepCommandHandler),
	}
}
