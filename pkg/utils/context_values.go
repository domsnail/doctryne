package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func GetClientDataFromIncomingMetadata(ctx context.Context) (id, from string) {
	cId := metadata.ValueFromIncomingContext(ctx, "client.id")
	if len(cId) != 0 {
		id = cId[0]
	}

	cFrom := metadata.ValueFromIncomingContext(ctx, "client.from")
	if len(cFrom) != 0 {
		from = cFrom[0]
	}

	return
}
