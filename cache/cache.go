package cache

import (
	"context"

	parser "github.com/levelsoftware/echoip/iputil/paser"
)

type CachedResponse struct {
	response *parser.Response
}

func (cr *CachedResponse) Build(response parser.Response) CachedResponse {
	return CachedResponse{
		response: &response,
	}
}

func (cr *CachedResponse) Get() parser.Response {
	return *cr.response
}

func (cr *CachedResponse) IsSet() bool {
	return cr.response != nil
}

type Cache interface {
	Get(ctx context.Context, ip string, cachedResponse *CachedResponse) error
	Set(ctx context.Context, ip string, response CachedResponse, cacheTtl int) error
}
