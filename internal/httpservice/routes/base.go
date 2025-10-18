package routes

import (
	"delegator/pkg/domain"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterBaseRoutes(
	router *gin.Engine,
	logger *slog.Logger,
	useCase domain.UseCase,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok"})
	})

	xtz := router.Group("/xtz")
	xtz.GET("/delegations", func(c *gin.Context) {
		res, err := useCase.GetDelegations(c)
		if err != nil {
			logger.Warn("failed to get delegations: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "failed to get delegations",
			})
			return
		}

		c.JSON(http.StatusOK, res)
	})
}

func CreateDelegatorRegistrar(
	logger *slog.Logger,
	queryUseCase domain.UseCase,
) RouteRegistrar {
	return func(engine *gin.Engine) {
		RegisterBaseRoutes(engine, logger, queryUseCase)
	}
}
