package types

import "context"

type Cache interface {
	Set(ctx context.Context, chatId int64, featureName string, stepName string) error
	Get(ctx context.Context, chatId int64) (*string, *string, error)
	Del(ctx context.Context, chatId int64) error

	SaveUserIdChatId(ctx context.Context, userId int64, chatId int64) error
	GetChatIdByUserId(ctx context.Context, userId int64) (*int64, error)
}
