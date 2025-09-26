package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dansen/jetgeo-go/internal/geo"
	"github.com/dansen/jetgeo-go/internal/httpx"
	"github.com/dansen/jetgeo-go/internal/util"
)

func main() {
	var (
		dataPath = flag.String("data", "", "geo data parent path")
		levelStr = flag.String("level", "", "min level: province|city|district")
		addr     = flag.String("addr", ":8080", "listen address")
	)
	flag.Parse()

	cfg := geo.LoadFromEnv(geo.DefaultConfig())
	if *dataPath != "" {
		cfg.GeoDataPath = *dataPath
	}
	if *levelStr != "" {
		if lv, err := geo.ParseLevel(*levelStr); err == nil {
			cfg.Level = lv
		}
	}

	logger := util.NewLogger()
	defer logger.Sync()

	engine, err := geo.NewEngine(cfg)
	if err != nil {
		log.Fatalf("init engine: %v", err)
	}

	h := &httpx.ReverseHandler{Engine: engine, Log: logger}
	r := httpx.NewRouter(h)

	srv := &http.Server{Addr: *addr, Handler: r}
	go func() {
		log.Printf("jetgeo-go listening on %s level=%s data=%s", *addr, cfg.Level.String(), cfg.GeoDataPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("server stopped")
}
