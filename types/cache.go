package types

import "context"

type Cache interface {
	Set(ctx context.Context, chatId int64, featureName string, stepName string) error
	Get(ctx context.Context, chatId int64) (*string, *string, error)
	Del(ctx context.Context, chatId int64) error
}
