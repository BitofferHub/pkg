package lock

import "context"

type Locker interface {
	Lock(ctx context.Context, key string) (Lock, error)
}

type Lock interface {
	Unlock(ctx context.Context) error
}
