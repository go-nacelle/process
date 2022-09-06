package process

import "context"

type healthKeyType struct{}

var healthKey = healthKeyType{}

func HealthWithContext(ctx context.Context, health *Health) context.Context {
	return context.WithValue(ctx, healthKey, health)
}

func HealthFromContext(ctx context.Context) *Health {
	if v, ok := ctx.Value(healthKey).(*Health); ok {
		return v
	}
	return nil
}
