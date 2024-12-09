package main

//go:generate openapi-generator-cli generate -g go-gin-server -i openapi.yaml -o openapi -p=interfaceOnly=true
//go:generate rm -r openapi/api
//go:generate rm openapi/main.go
//go:generate rm openapi/go.mod
//go:generate rm openapi/Dockerfile
