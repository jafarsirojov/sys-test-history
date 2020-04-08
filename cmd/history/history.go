package main

import (
	"context"
	"flag"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jafarsirojov/mux/pkg/mux"
	"github.com/jafarsirojov/sys-test-history/cmd/history/app"
	"github.com/jafarsirojov/sys-test-history/pkg/core/history"
	"log"
	"net"
	"net/http"
)

var (
	host = flag.String("host", "", "Server host")
	port = flag.String("port", "", "Server port")
	dsn  = flag.String("dsn", "", "Postgres DSN")
)

//-host 0.0.0.0 -port 9001 -dsn postgres://user:pass@localhost:5301/app
const ENV_PORT = "PORT"
const ENV_DSN = "DATABASE_URL"
const ENV_HOST = "HOST"

func main() {
	flag.Parse()
	envPort, ok := FromFlagOrEnv(*port, ENV_PORT)
	if !ok {
		log.Println("can't port")
		return
	}
	envDsn, ok := FromFlagOrEnv(*dsn, ENV_DSN)
	if !ok {
		log.Println("can't dsn")
		return
	}
	envHost, ok := FromFlagOrEnv(*host, ENV_HOST)
	if !ok {
		log.Println("can't host")
		return
	}
	addr := net.JoinHostPort(envHost, envPort)
	log.Println("starting server!")
	log.Printf("host = %s, port = %s\n", envHost, envPort)

	pool, err := pgxpool.Connect(
		context.Background(),
		envDsn,
	)
	if err != nil {
		panic(err)
	}
	historySvc := history.NewService(pool)
	historySvc.Start()
	exactMux := mux.NewExactMux()
	server := app.NewMainServer(exactMux, historySvc)
	exactMux.GET("/api/sys-test-history",
		server.HandleGetShowMarksUser,
		jwtMiddleware,
		requestIdier,
		logger,
	)
	exactMux.GET("/api/sys-test-history/user/{uid}",
		server.HandleGetShowMarksByUserIDIsAdmin,
		jwtMiddleware,
		requestIdier,
		logger,
	)
	exactMux.GET("/api/sys-test-history/group/{gid}",
		server.HandleGetShowMarksByGroupIDIsAdmin,
		jwtMiddleware,
		requestIdier,
		logger,
	)
	exactMux.POST("/api/sys-test-history",
		server.HandlePostAddHistory,
		jwtMiddleware,
		requestIdier,
		logger,
	)
	exactMux.DELETE("/api/sys-test-history/{id}",
		server.HandleDeleteRemovedHistoryByUserID,
		jwtMiddleware,
		requestIdier,
		logger,
	)
	panic(http.ListenAndServe(addr, server))
}
