// Package app provides self-contained application business logic and signal handling.
package app

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gevulotnetwork/devnet-explorer/api"
	"github.com/gevulotnetwork/devnet-explorer/signalhandler"
	"github.com/gevulotnetwork/devnet-explorer/store/cache"
	"github.com/gevulotnetwork/devnet-explorer/store/mock"
	"github.com/gevulotnetwork/devnet-explorer/store/pg"
)

// Run starts the application and listens for OS signals to gracefully shutdown.
func Run(args ...string) error {
	conf := ParseConfig(args...)
	s, err := createStore(conf)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	c := cache.New(s, conf.CacheRefreshInterval)
	srv, err := api.NewServer(conf.ServerListenAddr, c)
	if err != nil {
		return fmt.Errorf("failed to api server: %w", err)
	}

	sh := signalhandler.New(os.Interrupt)
	r := NewRunner(sh, srv, c)
	return r.Run()
}

func createStore(c Config) (api.Store, error) {
	if c.MockStore {
		return mock.New(), nil
	}
	return pg.New(c.DSN)
}

type Config struct {
	ServerListenAddr     string
	DSN                  string
	MockStore            bool
	CacheRefreshInterval time.Duration
}

// TODO: Proper config parsing
func ParseConfig(args ...string) Config {
	addr := os.Getenv("SERVER_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8383"
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "postgres://gevulot:gevulot@localhost:5432/gevulot"
	}

	cacheRefreshInterval := os.Getenv("CACHE_REFRESH_INTERVAL")
	if cacheRefreshInterval == "" {
		cacheRefreshInterval = "5s"
	}

	d, err := time.ParseDuration(cacheRefreshInterval)
	if err != nil {
		slog.Error("failed to parse cache refresh interval, defaulting to 5s", slog.Any("error", err))
		d = 5 * time.Second
	}

	mockStore, _ := strconv.ParseBool(os.Getenv("MOCK_STORE"))

	return Config{
		ServerListenAddr:     addr,
		DSN:                  dsn,
		MockStore:            mockStore,
		CacheRefreshInterval: d,
	}
}
