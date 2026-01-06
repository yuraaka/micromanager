// Package core implements the business logic for the chatter service,
// handling message and chat management with real-time notifications.
package core

import (
	"context"

	"{{snake .ServiceName}}/api"
)

type service struct {
	ctx *ServiceContext
}

var _ api.Service = (*service)(nil)

func NewServiceCore(ctx *ServiceContext) *Service {
	return &service{
		ctx: ctx,
	}
}

func (s *service) SayHello(ctx context.Context, req api.HelloRequest) (*api.HelloResponse, error) {
	return &api.HelloResponse{
		Message: "Hello, " + req.Name + ", " + s.ctx.Config.GreetingTail,
	}, nil
}
