package routes

import (
	"context"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/httpx"
	"github.com/querycap/pipeline/services/pipeline-serve/global"
)

var RootRouter = courier.NewRouter(httptransport.BasePath("/"))

func init() {
	RootRouter.Register(courier.NewRouter(&Next{}))
}

type Next struct {
	httpx.MethodPost `path:"next"`
}

func (n *Next) Output(ctx context.Context) (interface{}, error) {
	req := httptransport.HttpRequestFromContext(ctx)

	result, err := global.Pipeline.Next(ctx, req.Body)
	if err != nil {
		return nil, err
	}

	<-result.Done()
	if err := result.Err(); err != nil {
		return nil, err
	}

	for result.Scan() {
		return result.Next()
	}

	return nil, nil
}
