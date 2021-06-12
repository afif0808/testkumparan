package cache

import "context"

type StoreFunc func(ctx context.Context, keyword string, data interface{}) error
type GetFunc func(ctx context.Context, keyword string) (interface{}, error)
type ClearAllCacheFunc func(ctx context.Context) error
type GetKeysFunc func(ctx context.Context, pattern string) ([]string, error)
type DeleteFunc func(Ctx context.Context, keyword string) error
