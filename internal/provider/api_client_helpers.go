package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go"
	"os"
)

func getContextWithApiKey(ctx context.Context) context.Context {
	return context.WithValue(ctx, basistheory.ContextAPIKeys, map[string]basistheory.APIKey{
		"ApiKey": {Key: os.Getenv("BASISTHEORY_API_KEY")},
	})
}
