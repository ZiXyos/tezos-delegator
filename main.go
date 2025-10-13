package main

import (
	"context"
	"delegator/internal/core/delegator"
	"delegator/internal/http"
	"delegator/internal/http/routes"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/zixyos/glog"
	serviceloader "github.com/zixyos/goloader/service"
)

func main() {
	logger, err := glog.NewDefault()
	if err != nil {
		slog.New(
			slog.NewJSONHandler(os.Stdout, nil),
		).Error("failed to init logger", "error", err)
		os.Exit(84)
	}

	ctx := context.Background()

	engine := gin.New()

	httpServer := http.NewHTTPServer(
		http.WithEngine(engine),
		http.WithLogger(logger),
		http.WithHTTPServer(nil),
		http.WithRoutes(routes.RegisterBaseRoutes),
	)

	delegatorService := delegator.NewDelegator(
		delegator.WithLogger(logger),
		delegator.WithComponents(httpServer),
	)

	app := serviceloader.New(
		serviceloader.WithLogger(logger),
		serviceloader.WithService(delegatorService),
	)

	app.Run(ctx)
}
