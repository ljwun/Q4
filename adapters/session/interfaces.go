//go:generate mockgen -package=session -destination=mock.go -source=interfaces.go

package session

import "context"

type IStore interface {
	Load(ctx context.Context, name string) (map[string]string, error)
	Save(ctx context.Context, name string, data map[string]string) error
}

type ISession interface {
	Load() error
	Get(key string) string
	Set(key, value string)
	Delete(key string)
	Clear()
	Save() error
}
