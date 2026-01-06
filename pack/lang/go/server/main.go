package main

import (
	"context"

	logrus "github.com/sirupsen/logrus"
	"{{snake .ServiceName}}/core"
)

func main() {
	cfg := core.Config{
		GreetingTail: common.RequireEnv("GREETING_TAIL"),
	}

	ctx := context.Background()
	// todo: add database support with migration
	service := core.NewServiceCore(&core.ServiceContext{
		Config: &cfg,
	})

	router := NewRouter(service)
	if err := router.Run(); err != nil {
		logrus.WithError(err).Fatal("Failed to run router")
	}
}
