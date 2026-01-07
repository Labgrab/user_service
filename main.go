package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"labgrab/user_service/api/proto"
	"labgrab/user_service/internal/repository/sqlc"
	"labgrab/user_service/internal/service"
	"labgrab/user_service/pkg/config"
	"labgrab/user_service/pkg/logger"
	"labgrab/user_service/pkg/telemetry"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to parse .env: %v", err)
	}

	tp, err := telemetry.InitTracer(&telemetry.Config{
		ServiceName:    cfg.ServiceName,
		JaegerEndpoint: cfg.JaegerEndpoint,
		Environment:    string(cfg.Environment),
	})

	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}

	zapLogger := logger.Logger(&logger.Options{
		Path:       "logs/user-service.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})
	defer zapLogger.Sync()

	pgconfig, err := pgxpool.ParseConfig(cfg.DBConn)
	if err != nil {
		log.Fatalf("Failed to parse DB connection string: %v", err)
	}

	pgconfig.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithTrimSQLInSpanName(),
	)

	conn, err := pgxpool.NewWithConfig(ctx, pgconfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close()

	repo := sqlc.New(conn)
	svc := &service.Service{
		Logger: zapLogger,
		Repo:   repo,
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
		log.Println("Shutting down...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := telemetry.Shutdown(shutdownCtx, tp); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
		s.GracefulStop()
	}()

	log.Printf("Server started on port %d with tracing enabled", cfg.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
