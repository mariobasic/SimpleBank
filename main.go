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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}
	conn, err := sql.Open(config.DB.Driver, config.DB.Source)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	runDBMigration(config.DB.MigrationURL, config.DB.Source)

	store := db.NewStore(conn)
	//runGinServer(config, store) // left to show example of standalone gin server
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
}

func runGrpcServer(config util.Config, store db.Store) {
	server := gapi.NewServer(config, store)
	grpcServer := grpc.NewServer()
	//pb.RegisterSimpleBankServer(grpcServer, &pb.UnimplementedSimpleBankServer{})
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.Server.Grpc)
	if err != nil {
		log.Fatal("grpc server failed to listen:", err)
	}
	log.Println("grpc server listening on", config.Server.Grpc)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("grpc server failed to serve:", err)
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
		log.Fatal("cannot register gateway server handler:", err)
	}

	swagFs, err := fs.New()
	if err != nil {
		log.Fatalf("cannot create statik fs: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(swagFs)))

	listen, err := net.Listen("tcp", config.Server.Http)
	if err != nil {
		log.Fatal("cannot create listener", err)
	}
	log.Printf("start HTTP gateway server at %s", listen.Addr().String())
	err = http.Serve(listen, mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server:", err)
	}
}

func runGinServer(config util.Config, store db.Store) {
	server := api.NewServer(config, store)
	err := server.Start(config.Server.Http)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatalf("cannot create new migration instance: %s", err)
	}

	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to run migrate up: %s", err)
	}

	log.Println("migrate up successfully")
}
