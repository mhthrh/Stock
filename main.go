package main

import (
	"Github.com/mhthrh/Stock/API"
	"Github.com/mhthrh/Stock/Utilitys/ConfigUtil"
	"Github.com/mhthrh/Stock/Utilitys/DbUtil/DbPool"
	"Github.com/mhthrh/Stock/Utilitys/LogUtil"
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	ip   string
	port int
	cnn  = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"127.0.0.1", 5432, "postgres", "123456", "Stock")
)

func main() {

	flag.StringVar(&ip, "ip", "localhost", "What is listener IP address")
	flag.IntVar(&port, "port", 9999, "Port Number")
	flag.Parse()
	cfg := ConfigUtil.ReadConfig("Config/ConfigPlane.json")
	if cfg == nil {
		log.Fatalln("Cant read Config,By")
	}
	db := DbPool.New(&DbPool.DbInfo{
		Host:            cfg.DB[0].Host,
		Port:            cfg.DB[0].Port,
		User:            cfg.DB[0].User.UserName,
		Pass:            cfg.DB[0].User.Password,
		Dbname:          cfg.DB[0].Dbname,
		Driver:          cfg.DB[0].Driver,
		ConnectionCount: 10, // connection pool count
		RefreshPeriod:   20, // refresh time for checking connection health!
	})
	logger := LogUtil.New()
	sm := mux.NewRouter()

	API.RunApiOnRouter(sm, logger, db)

	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", ip, port),
		Handler:      sm,
		ErrorLog:     log.New(LogUtil.LogrusErrorWriter{}, "", 0),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  180 * time.Second,
	}

	go func() {
		fmt.Printf("Starting server on  %s:%d\n", ip, port)
		err := server.ListenAndServe()
		if err != nil {
			logger.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	log.Println("Got signal:", <-c)

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	server.Shutdown(ctx)
}
