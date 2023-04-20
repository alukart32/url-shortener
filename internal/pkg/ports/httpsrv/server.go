// Package httpsrv provides a HTTP server implementation.
package httpsrv

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alukart32/shortener-url/internal/pkg/zerologx"
	"github.com/caarlos0/env/v7"
)

// A server defines wrapper for the HTTP server.
type server struct {
	srv             *http.Server
	notify          chan error
	shutdownTimeout time.Duration
}

// Server returns the Server and starts it.
func Server(cfg Config, h http.Handler) (*server, error) {
	if len(cfg.ADDR) == 0 || cfg.Empty() {
		opts := env.Options{RequiredIfNoDef: true}
		if err := env.Parse(&cfg, opts); err != nil {
			return nil, fmt.Errorf("failed to read config: %v", err)
		}
	}

	tlsConf := newTLSConf(cfg.EnableHTTPS)

	srv := &http.Server{
		Addr:         cfg.ADDR,
		Handler:      h,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		TLSConfig:    tlsConf,
	}

	s := server{
		srv:             srv,
		notify:          make(chan error, 1),
		shutdownTimeout: cfg.ShutdownTimeout,
	}

	go func() {
		if cfg.EnableHTTPS {
			s.notify <- s.srv.ListenAndServeTLS("", "")
		} else {
			s.notify <- s.srv.ListenAndServe()
		}
		close(s.notify)
	}()

	return &s, nil
}

// Notify throws a server error.
func (s *server) Notify() <-chan error {
	return s.notify
}

// Shutdown gracefully stops the server during timeout.
func (s *server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.srv.Shutdown(ctx)
}

// Config represents the http server configuration.
type Config struct {
	//Port specifies the listening port of the server.
	ADDR string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`

	// EnableHTTPS indicates the TLS connection mode, such as tls, mtls or insecure (not set).
	EnableHTTPS bool `env:"ENABLE_HTTPS" envDefault:""`

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"300ms"`

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"300ms"`

	// ShutdownTimeout is the maximum duration before timing out
	// stops the running server. A zero or negative value means
	// there will be no timeout.
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"1s"`
}

// Empty checks on being empty.
func (c Config) Empty() bool {
	return len(c.ADDR) == 0 &&
		!c.EnableHTTPS &&
		c.ReadTimeout == 0 &&
		c.WriteTimeout == 0 &&
		c.ShutdownTimeout == 0
}

func newTLSConf(isOn bool) *tls.Config {
	var conf *tls.Config
	logger := zerologx.Get()

	if isOn {
		logger.Info().Msg("http: TLS is on")
		// Load CA cert.
		caCert, err := os.ReadFile(os.Getenv("CA_CERT"))
		if err != nil {
			logger.Error().Err(err).Msg("read ca cert")
			panic(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		conf = &tls.Config{
			RootCAs: caCertPool,
		}
	} else {
		logger.Info().Msg("http: TLS is off")
		conf = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return conf
}
