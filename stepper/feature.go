package stepper

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	tbs "github.com/lalkalol1907/tg-bot-stepper"
)

type Feature struct {
	steps map[string]tbs.Step

	firstStep string
}

func (f *Feature) AddStep(name string, step tbs.Step) *Feature {
	f.steps[name] = step
	if len(f.steps) == 1 {
		f.firstStep = name
	}

	return f
}

func (f *Feature) Run(ctx context.Context, stepName string, b *bot.Bot, update *models.Update) (tbs.StepExecutionResult, error) {
	if len(stepName) == 0 {
		stepName = f.firstStep
	}

	response, err := f.steps[stepName](ctx, b, update)
	if err != nil {
		return response, err
	}

	return response, nil
}

func NewFeature() *Feature {
	return &Feature{
		steps: make(map[string]tbs.Step),
	}
}
