package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/0x24CaptainParrot/gomarket/internal/config"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/handlers"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gomarket/internal/pkg/services"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := repository.GetPostgresDB(cfg.DBUri)
	if err != nil {
		log.Fatal(err)
	}

	repository := repository.NewRepository(db)
	services := services.NewService(services.Dependencies{
		AuthService: services.NewAuthService(repository.AuthorizationRepository, cfg),
	})
	handler := handlers.NewHandler(services, cfg)

	srv := &handlers.Server{}
	go func() {
		if err := srv.Run(cfg.RunAddr, handler.InitApiRoutes()); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("server is shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("error occured on server while shutting down: %v", err.Error())
	}
}
