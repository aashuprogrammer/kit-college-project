package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aashuprogrammer/fee-management-system/api"
	"github.com/aashuprogrammer/fee-management-system/db/pgdb"
	"github.com/aashuprogrammer/fee-management-system/token"
	"github.com/aashuprogrammer/fee-management-system/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// config
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("failed to read env", err)
	}

	// db
	poolConfig, err := pgxpool.ParseConfig(config.DBUrl)
	if err != nil {
		log.Fatalln("Unable to parse DATABASE_URL:", err)
	}

	dbConn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}
	store := pgdb.NewStore(dbConn)
	defer dbConn.Close()

	// token
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmectricKey)
	if err != nil {
		log.Fatal("failed to create token maker ", err)
	}
	fmt.Println("apple")
	// fiber server
	server, err := api.NewServer(config, store, tokenMaker)
	fmt.Println(server)

	if err != nil {
		log.Fatal("failed to create server ", err)
	}
	err = server.Start(config.Port)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}
