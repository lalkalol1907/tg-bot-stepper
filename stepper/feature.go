package stepper

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/lalkalol1907/tg-bot-stepper/types"
)

type Feature struct {
	steps map[string]types.Step

	firstStep string
}

func (f *Feature) AddStep(name string, step types.Step) *Feature {
	f.steps[name] = step
	if len(f.steps) == 1 {
		f.firstStep = name
	}

	return f
}

func (f *Feature) Run(ctx context.Context, stepName *string, b *bot.Bot, update *models.Update) (types.StepExecutionResult, error) {
	if stepName == nil {
		stepName = &f.firstStep
	}

	response, err := f.steps[*stepName](ctx, b, update)
	if err != nil {
		return response, err
	}

	return response, nil
}

func NewFeature() *Feature {
	return &Feature{
		steps: make(map[string]types.Step),
	}
}
