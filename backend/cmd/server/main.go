package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"smart-inventory-backend/internal/config"
	"smart-inventory-backend/internal/db"
	"smart-inventory-backend/internal/handlers"
	"smart-inventory-backend/internal/repo"
	"smart-inventory-backend/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatal(err)
		}
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
		ginWriter := mw
		handlers.SetGinWriters(ginWriter)
	}

	writePool, err := db.NewPool(ctx, cfg.WriteDBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer writePool.Close()
	logDBInfo(ctx, writePool, "write-db")

	var readPool = writePool
	if cfg.EnableReplica {
		readPool, err = db.NewPool(ctx, cfg.ReadDBURL)
		if err != nil {
			log.Fatal(err)
		}
		defer readPool.Close()
		logDBInfo(ctx, readPool, "read-db")
	}

	pgRepo := repo.NewPgRepo(writePool, readPool)
	svc := service.New(writePool, pgRepo, pgRepo, pgRepo, pgRepo, pgRepo)
	h := handlers.New(svc, cfg.APIKey)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: h.Router(),
	}

	go func() {
		log.Printf("HTTP listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func logDBInfo(ctx context.Context, pool *pgxpool.Pool, label string) {
	var user, db string
	err := pool.QueryRow(ctx, `SELECT current_user, current_database()`).Scan(&user, &db)
	if err != nil {
		log.Printf("%s info error: %v", label, err)
		return
	}
	log.Printf("%s connected user=%s db=%s", label, user, db)
}
