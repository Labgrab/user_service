package main

import (
	"context"
	"labgrab/user_service/api/proto"
	"labgrab/user_service/internal/repository/sqlc"
	"labgrab/user_service/internal/service"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conn, err := pgx.ConnectConfig(pgx.ConnConfig{})
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	repo := sqlc.New(conn)
	svc := &service.Service{
		Repo: repo,
	}

	lis, err := net.Listen("tcp", ":5051")
	if err != nil {
		log.Fatalf("failed to listen on port 5051: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterUserServiceServer(s, svc)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Println("server started")
	<-ctx.Done()
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer timeoutCancel()

	s.GracefulStop()
	if err := conn.Close(timeoutCtx); err != nil {
		log.Fatalf("failed to close DB connection: %v", err)
	}
}
