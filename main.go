package main

import (
	"github.com/gin-gonic/gin"

	"q4/api"
	"q4/api/openapi"
)

func main() {
	args := ParseArgs()
	if !args.Validate() {
		panic("missing arguments")
	}
	strictServer, err := api.NewServer(args.ServerConfig)
	if err != nil {
		panic(err)
	}
	defer strictServer.Close()

	router := gin.Default()
	handler := openapi.NewStrictHandler(strictServer, nil)
	openapi.RegisterHandlers(router, handler)
	if err := router.Run(args.ServerURL); err != nil {
		panic(err)
	}
}
