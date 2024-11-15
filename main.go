package main

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mariobasic/simplebank/api"
	db "github.com/mariobasic/simplebank/db/sqlc"
	_ "github.com/mariobasic/simplebank/doc/statik"
	"github.com/mariobasic/simplebank/gapi"
	"github.com/mariobasic/simplebank/mail"
	"github.com/mariobasic/simplebank/pb"
	"github.com/mariobasic/simplebank/util"
	"github.com/mariobasic/simplebank/worker"
	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msgf("Error loading config: %s", err)
	}

	if config.Env == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	pool, err := pgxpool.New(ctx, config.DB.Source)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: %s", err)
	}

	runDBMigration(config.DB.MigrationURL, config.DB.Source)

	store := db.NewStore(pool)

	redisOpt := asynq.RedisClientOpt{Addr: config.Server.Redis}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	waitGroup, ctx := errgroup.WithContext(ctx)

	//runGinServer(config, store) // left to show example of standalone gin server
	runTaskProcessor(ctx, waitGroup, redisOpt, store, config)
	runGatewayServer(ctx, waitGroup, config, store, taskDistributor)
	runGrpcServer(ctx, waitGroup, config, store, taskDistributor)
	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
}

func runGrpcServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
	server := gapi.NewServer(config, store, taskDistributor)

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)

	//pb.RegisterSimpleBankServer(grpcServer, &pb.UnimplementedSimpleBankServer{})
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.Server.Grpc)
	if err != nil {
		log.Fatal().Msgf("grpc server failed to listen:  %s", err)
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("grpc server listening on: %s", config.Server.Grpc)
		err = grpcServer.Serve(listener)
		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Error().Err(err).Msg("grpc server failed to serve")
			return err
		}
		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("grpc server gracefully shutting down")
		grpcServer.GracefulStop()
		log.Info().Msg("grpc server is stopped")
		return nil
	})

}

func runGatewayServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
	server := gapi.NewServer(config, store, taskDistributor)

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}))

	err := pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msgf("cannot register gateway server handler: %s", err)
	}

	swagFs, err := fs.New()
	if err != nil {
		log.Fatal().Msgf("cannot create statik fs: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(swagFs)))

	c := cors.Default()
	handler := c.Handler(gapi.HttpLogger(mux))

	httpServer := &http.Server{
		Handler: handler,
		Addr:    config.Server.Http,
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start HTTP gateway server at %s", httpServer.Addr)
		err = httpServer.ListenAndServe()
		if err != nil {
			log.Err(err).Msg("cannot start HTTP gateway server")
			return err
		}
		return nil
	})
	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("http gateway server gracefully shutting down")
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			log.Error().Err(err).Msg("cannot shutdown HTTP gateway server")
			return err
		}
		log.Info().Msg("http gateway server is stopped")
		return nil
	})
}

//goland:noinspection GoUnusedFunction
func runGinServer(config util.Config, store db.Store) {
	server := api.NewServer(config, store)
	err := server.Start(config.Server.Http)
	if err != nil {
		log.Fatal().Msgf("cannot start server: %s", err)
	}
}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msgf("cannot create new migration instance: %s", err)
	}

	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Msgf("failed to run migrate up: %s", err)
	}

	log.Info().Msg("migrate up successfully")
}

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	config util.Config,
) {
	sender := mail.NewGmailSender(config.Email.Sender.Name, config.Email.Sender.Address, config.Email.Sender.Password)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, sender)
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("task processor failed to start")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("task processor gracefully shutting down")
		taskProcessor.Shutdown()
		log.Info().Msg("task processor is stopped")
		return nil
	})
}
