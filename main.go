package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/mariobasic/simplebank/api"
	db "github.com/mariobasic/simplebank/db/sqlc"
	_ "github.com/mariobasic/simplebank/doc/statik"
	"github.com/mariobasic/simplebank/gapi"
	"github.com/mariobasic/simplebank/pb"
	"github.com/mariobasic/simplebank/util"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msgf("Error loading config: %s", err)
	}

	if config.Env == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DB.Driver, config.DB.Source)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: %s", err)
	}

	runDBMigration(config.DB.MigrationURL, config.DB.Source)

	store := db.NewStore(conn)
	//runGinServer(config, store) // left to show example of standalone gin server
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
}

func runGrpcServer(config util.Config, store db.Store) {
	server := gapi.NewServer(config, store)

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)

	//pb.RegisterSimpleBankServer(grpcServer, &pb.UnimplementedSimpleBankServer{})
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.Server.Grpc)
	if err != nil {
		log.Fatal().Msgf("grpc server failed to listen:  %s", err)
	}
	log.Info().Msgf("grpc server listening on: %s", config.Server.Grpc)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msgf("grpc server failed to serve: %s", err)
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server := gapi.NewServer(config, store)

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	listen, err := net.Listen("tcp", config.Server.Http)
	if err != nil {
		log.Fatal().Msgf("cannot create listener: %s", err)
	}
	log.Info().Msgf("start HTTP gateway server at %s", listen.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listen, handler)
	if err != nil {
		log.Fatal().Msgf("cannot start HTTP gateway server: %s", err)
	}
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
