package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/GFernandesS/go-bid/internal/api"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func init() {
	gob.Register(uuid.UUID{})
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	apiRouter, err := api.GetApi(ctx)

	if err != nil {
		panic(err)
	}

	defer apiRouter.Close()

	apiRouter.BindRoutes()

	fmt.Println("Start server on port 8080")

	if err := http.ListenAndServe(":8080", apiRouter.Router); err != nil {
		panic(err)
	}
}
