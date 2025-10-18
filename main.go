package main

import (
	"context"
	"database/sql"
	"delegator/conf"
	"delegator/internal/core/delegator"
	"delegator/internal/core/indexer"
	"delegator/internal/database"
	"delegator/internal/httpservice"
	"delegator/internal/httpservice/routes"
	"delegator/internal/services"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/zixyos/glog"
	serviceloader "github.com/zixyos/goloader/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed database/sql/*.sql
var migrationFS embed.FS

func buildConnectionString(config *conf.DelegatorConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Storage.Database.Host,
		config.Storage.Database.Port,
		config.Storage.Database.Username,
		config.Storage.Database.Password,
		config.Storage.Database.Database,
	)
}

func main() {
	logger, err := glog.NewDefault()
	if err != nil {
		slog.New(
			slog.NewJSONHandler(os.Stdout, nil),
		).Error("failed to init logger", "error", err)
		os.Exit(84)
	}

	ctx := context.Background()

	delegatorConf, err := conf.LoadConfig()
	if err != nil {
		logger.Warn("failed to load delegator config", "error", err)
		os.Exit(84)
	}
	logger.Info(fmt.Sprintf("%+v", delegatorConf))

	connectionString := buildConnectionString(delegatorConf)
	logger.Info(fmt.Sprintf("connecting to postgres at %s", connectionString))
	dbDriver, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Warn("failed to open database connection", "error", err)
		os.Exit(84)
	}

	pgClient, err := database.NewClient(
		ctx,
		database.WithLogger(logger),
		database.WithDriver(dbDriver),
	)

	if err != nil {
		logger.Warn("failed to init database client", "error", err)
		os.Exit(84)
	}

	_, err = gorm.Open(postgres.New(postgres.Config{
		Conn: pgClient.Driver,
	}))

	if err != nil {
		logger.Warn("failed to init database client", "error", err)
		os.Exit(84)
	}

	err = database.RunMigrations(pgClient.Driver, migrationFS)
	if err != nil {
		logger.Warn("failed to run migrations", "error", err)
		os.Exit(84)
	}

	delegatorRepository := delegator.NewRepository(
		delegator.RepositoryWithLogger(logger),
		delegator.RepositoryWithDBClient(dbDriver),
	)

	delegatorUseCase := delegator.NewUseCase(
		delegator.UseCaseWithLogger(logger),
		delegator.UseCaseWithRepository(delegatorRepository),
	)

	engine := gin.New()
	httpClient := &http.Client{Timeout: time.Duration(delegatorConf.HTTP.ReadTimeout) * time.Second}

	httpServer := httpservice.NewHTTPServer(
		httpservice.WithEngine(engine),
		httpservice.WithLogger(logger),
		httpservice.WithHTTPServer(delegatorConf),
		httpservice.WithRoutes(
			routes.CreateDelegatorRegistrar(logger, delegatorUseCase),
		),
	)

	tzktHTTPHandler := services.NewHTTPHandler(
		services.HandlerWithLogger(logger),
		services.HandlerWithClient(httpClient),
		services.HandlerWithBaseURL("https://api.tzkt.io/v1/"), // TODO: Add tzkt config
	)

	indexerComponent := indexer.NewDelegatorIndexer(
		indexer.WithLogger(logger),
		indexer.WithDelegationHandler(tzktHTTPHandler),
	)

	delegatorService := delegator.NewDelegator(
		delegator.WithLogger(logger),
		delegator.WithComponents(httpServer, indexerComponent),
	)

	app := serviceloader.New(
		serviceloader.WithLogger(logger),
		serviceloader.WithService(delegatorService),
	)

	app.Run(ctx)
}
