package shortener

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alukart32/shortener-url/internal/pkg/db/migrate"
	dbpgx "github.com/alukart32/shortener-url/internal/pkg/db/postgres"
	"github.com/alukart32/shortener-url/internal/pkg/ginx"
	"github.com/alukart32/shortener-url/internal/pkg/jwttoken"
	"github.com/alukart32/shortener-url/internal/pkg/middleware/trustsubnet"
	"github.com/alukart32/shortener-url/internal/pkg/ports/grpcauth"
	"github.com/alukart32/shortener-url/internal/pkg/ports/grpcsrv"
	"github.com/alukart32/shortener-url/internal/pkg/ports/httpauth"
	"github.com/alukart32/shortener-url/internal/pkg/ports/httpsrv"
	"github.com/alukart32/shortener-url/internal/pkg/zerologx"
	grpcv1 "github.com/alukart32/shortener-url/internal/shortener/controller/grpc/v1"
	"github.com/alukart32/shortener-url/internal/shortener/controller/http/pinger"
	httpv1 "github.com/alukart32/shortener-url/internal/shortener/controller/http/v1"
	"github.com/alukart32/shortener-url/internal/shortener/services"
	"github.com/alukart32/shortener-url/internal/shortener/services/pingpgx"
	"github.com/alukart32/shortener-url/internal/shortener/services/shorturl"
	"github.com/alukart32/shortener-url/internal/shortener/storage/shortenedurl/filestorage"
	"github.com/alukart32/shortener-url/internal/shortener/storage/shortenedurl/memstorage"
	shortenedurlpgx "github.com/alukart32/shortener-url/internal/shortener/storage/shortenedurl/postgres"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run starts shortener app.
func Run() {
	logger := zerologx.Get()

	conf, err := prepareConf()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to read app config")
	}

	var pgxPool *pgxpool.Pool
	if len(conf.DatabaseDSN) != 0 {
		logger.Info().Msg("prepare: postgres pool")
		pgxPool, err = dbpgx.Get(conf.DatabaseDSN)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to prepare: postgres pool")
		}
		defer func() {
			logger.Info().Msg("shutdown: postgres pool")
			pgxPool.Close()
		}()

		logger.Info().Msg("start db migration")
		if err = migrate.Up(conf.DatabaseDSN, ""); err != nil {
			logger.Fatal().Err(err).Msg("failed to migrate db")
		}
	}

	//Prepare services.
	logger.Info().Msg("prepare: services")
	servs, shutdown := prepareServices(conf, pgxPool)
	defer func() {
		if shutdown != nil {
			if err := shutdown(); err != nil {
				logger.Error().Err(err).Msg("failed to shutdown services")
			}
		}
	}()

	// Prepare HTTP router.
	logger.Info().Msg("prepare: gin router")
	ginRouter, err := ginx.Get()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to prepare: gin")
	}

	setHTTPRoutes(conf, ginRouter, servs)

	// Add postgres ping handler.
	pinger.PostgresPinger(ginRouter, pingpgx.Pinger(pgxPool, 1*time.Second))

	// Add HTTP profiler.
	if len(os.Getenv("PPROF_ON")) != 0 {
		logger.Info().Msg("enable: http pprof")
		pprof.Register(ginRouter)
	}

	// Run http server.
	logger.Info().Msg("run: http server")
	httpsrvConf := httpsrv.Config{
		ADDR:        conf.Addr,
		EnableHTTPS: conf.EnableHTTPS,
	}
	httpServer, err := httpsrv.Server(httpsrvConf, ginRouter)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	defer func() {
		logger.Info().Msg("shutdown: http server")
		if err = httpServer.Shutdown(); err != nil {
			logger.Error().Err(err).Send()
		}
	}()

	// Prepare grpc server.
	logger.Info().Msg("prepare: grpc server")
	jwtManager, err := jwttoken.Manager(
		jwttoken.Config{
			Key:     conf.JWTSignKey,
			ExpTime: conf.JWTExpTime,
		})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to prepare: JWT manager")
	}

	grpcServer, err := grpcsrv.Server(
		grpcsrv.Config{
			ADDR:      conf.GrpcAddr,
			EnableTLS: conf.EnableHTTPS,
		}, grpcauth.NewAuthOpts(
			jwtManager,
			grpcv1.MethodsForAuthSkip(),
		),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to prepare: grpc server")
	}

	// Run grpc server.
	logger.Info().Msg("run: grpc server")
	grpcv1.RegServices(grpcServer.Srv, servs)
	grpcServer.Run()
	defer func() {
		logger.Info().Msg("shutdown: grpc server")
		grpcServer.Shutdown()
	}()

	// Waiting signals.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case s := <-interrupt:
		logger.Info().Msg(s.String())
	case err = <-httpServer.Notify():
		logger.Fatal().Err(err).Msg("failed to run http server")
	case err = <-grpcServer.Notify():
		logger.Fatal().Err(err).Msg("failed to run grpc server")
	}
}

// shutdownFn defines a func that can shut down anything.
type shutdownFn func() error

// prepareServices prepares services.
func prepareServices(conf config, pgxPool *pgxpool.Pool) (*services.Services, shutdownFn) {
	logger := zerologx.Get()

	var (
		shortener services.Shortener
		provider  services.Provider
		deleter   services.Deleter
		statistic services.StatProvider
		pinger    services.Pinger

		err      error
		shutdown shutdownFn
	)

	if pgxPool != nil {
		shortener, err = shorturl.Shortener(
			conf.BaseURL,
			shortenedurlpgx.ShortURLSaver(pgxPool),
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to prepare url shortener")
		}
		provider = shortenedurlpgx.ShortURLProvider(pgxPool)
		statistic = shortenedurlpgx.StatProvider(pgxPool)
		deleter = shortenedurlpgx.ShortURLDeleter(pgxPool)
		pinger = pingpgx.Pinger(pgxPool, 1*time.Second)
	}

	if len(conf.FileStoragePath) != 0 {
		fileStorage, err := filestorage.FileStorage(conf.FileStoragePath)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to preapre file storage")
		}
		shutdown = func() error {
			return fileStorage.Close()
		}

		shortener, err = shorturl.Shortener(conf.BaseURL, fileStorage)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to prepare url shortener")
		}
		provider = fileStorage
		statistic = fileStorage
		deleter = fileStorage
	}

	if pgxPool == nil && len(conf.FileStoragePath) == 0 {
		memStorage := memstorage.MemStorage()

		shortener, err = shorturl.Shortener(conf.BaseURL, memStorage)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to prepare url shortener")
		}
		provider = memStorage
		statistic = memStorage
		deleter = memStorage
	}

	servs := services.NewServices(shortener, provider, deleter, statistic, pinger)
	return servs, shutdown
}

// setHTTPRoutes adds HTTP routes to the gin router.
func setHTTPRoutes(conf config, g *gin.Engine, servs *services.Services) {
	logger := zerologx.Get()

	cookieAuth, err := httpauth.CookieAuthProvider()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to prepare cookie auth provider")
	}

	subnetValidator, err := trustsubnet.Validator(conf.TrustedSubnet)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to prepare subnet validator")
	}

	httpv1.SetRoutes(g, cookieAuth, subnetValidator, servs)
}

// config is the representation of shortener app settings.
type config struct {
	Addr            string        `json:"server_address"`
	GrpcAddr        string        `json:"grpc_server_address"`
	BaseURL         string        `json:"base_url"`
	FileStoragePath string        `json:"file_storage_path"`
	DatabaseDSN     string        `json:"database_dsn"`
	EnableHTTPS     bool          `json:"enable_https"`
	TrustedSubnet   string        `json:"trusted_subnet"`
	JWTSignKey      string        `json:"jwt_sign_key"`
	JWTExpTime      time.Duration `json:"jwt_token_expr"`
}

// prepareConf prepres shortener app config.
func prepareConf() (config, error) {
	var (
		confFile string
		conf     config
	)

	flag.StringVar(&conf.Addr, "a", "", "server address")
	flag.StringVar(&conf.BaseURL, "b", "", "baseURL for shortening")
	flag.StringVar(&conf.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&conf.DatabaseDSN, "d", "", "postgres DSN")
	flag.BoolVar(&conf.EnableHTTPS, "s", false, "server tls mode")
	flag.StringVar(&conf.TrustedSubnet, "t", "", "trusted subnet")
	flag.StringVar(&confFile, "c", "", "configuration filepath")
	flag.StringVar(&confFile, "config", "", "configuration filepath")
	flag.Parse()

	envConfigFile := os.Getenv("CONFIG")
	if len(confFile) == 0 && len(envConfigFile) == 0 {
		return conf, nil
	}

	if len(confFile) == 0 && len(envConfigFile) != 0 {
		confFile = envConfigFile
	}

	b, err := os.ReadFile(confFile)
	if err != nil {
		return config{}, fmt.Errorf("failed to read config file")
	}

	if err = json.Unmarshal(b, &conf); err != nil {
		return config{}, fmt.Errorf("failed to unmarshal config file")
	}

	return conf, nil
}
