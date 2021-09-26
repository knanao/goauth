package main

import (
	"context"
	"html/template"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/knanao/goauth/server/model"
	"github.com/knanao/goauth/server/session"
	"github.com/knanao/goauth/server/setting"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

var (
	templates      map[string]*template.Template
	sessionManager *session.Manager
	userDA         *model.UserDataAccessor
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)

	t := &Template{}
	e.Renderer = t

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	setStaticRoute(e)
	setRoute(e)

	sessionManager = &session.Manager{}
	sessionManager.Start(e)

	userDA = &model.UserDataAccessor{}
	userDA.Start(e)

	go func() {
		if err := e.Start(setting.Server.Port); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Info(err)
		e.Close()
	}

	userDA.Stop()

	sessionManager.Stop()

	time.Sleep(1 * time.Second)
}

func init() {
	setting.Load()
	loadTemplates()
}
