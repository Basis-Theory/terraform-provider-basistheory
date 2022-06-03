package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go/v2"
	"os"
)

func getContextWithApiKey(ctx context.Context, apiKey string) context.Context {
	if apiKey != "" {
		return context.WithValue(ctx, basistheory.ContextAPIKeys, map[string]basistheory.APIKey{
			"ApiKey": {Key: apiKey},
		})
	}

	return context.WithValue(ctx, basistheory.ContextAPIKeys, map[string]basistheory.APIKey{
		"ApiKey": {Key: os.Getenv("BASISTHEORY_API_KEY")},
	})
}
