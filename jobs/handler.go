package jobs

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/tuyentv96/hasty-challenge/config"
)

type HTTPHandler struct {
	config  config.Config
	routes  *echo.Echo
	service Service
}

func NewHTTPHandler(cfg config.Config, svc Service) *HTTPHandler {
	h := HTTPHandler{
		config:  cfg,
		service: svc,
	}

	h.InitRoutes()
	return &h
}

func (a *HTTPHandler) InitRoutes() {
	a.routes = echo.New()
	a.routes.HideBanner = true
	a.routes.Use(middleware.Recover())

	if a.config.HTTPLogger {
		a.routes.Use(middleware.Logger())
	}

	a.routes.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	a.routes.GET("/health", func(context echo.Context) error {
		return context.String(http.StatusOK, "OK")
	})

	v1 := a.routes.Group("/v1")
	v1.POST("/jobs", a.SaveJobHandler)

	jobs := v1.Group("/jobs")
	jobs.GET("/:id", a.GetJobHandler)
}

func (a *HTTPHandler) Serve() {
	a.routes.Start(fmt.Sprintf(":%d", a.config.HTTPConfig.HTTPPort))
}

func (a *HTTPHandler) GetJobHandler(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse id")
	}

	result, err := a.service.GetJobByID(ctx.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrJobNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, ErrJobNotFound.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, result)
}

func (a *HTTPHandler) SaveJobHandler(ctx echo.Context) error {
	var job JobPayload
	if err := ctx.Bind(&job); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := a.service.SaveJob(ctx.Request().Context(), job)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusCreated, result)
}
