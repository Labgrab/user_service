package main

import (
	"context"
	"fmt"
	"labgrab/user_service/api/proto"
	"labgrab/user_service/internal/repository/sqlc"
	"labgrab/user_service/internal/service"
	"labgrab/user_service/pkg/config"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to parse .env: %v", err)
	}

	pgconfig, err := pgxpool.ParseConfig(cfg.DBConn)
	if err != nil {
		log.Fatalf("Failed to parse DB connection string: %v", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgconfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	repo := sqlc.New(conn)
	svc := &service.Service{
		Repo: repo,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	proto.RegisterUserServiceServer(s, svc)

	go func() {
		<-ctx.Done()
		s.GracefulStop()
		conn.Close()
	}()

	log.Println("server started")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
