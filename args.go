package main

import "github.com/spf13/pflag"

type Args struct {
	RedisURL  string
	NatsURL   string
	ServerURL string
}

func ParseArgs() Args {
	args := Args{}
	pflag.StringVar(&args.RedisURL, "redis-url", "localhost:6379", "")
	pflag.StringVar(&args.NatsURL, "nats-url", "localhost:4222", "")
	pflag.StringVar(&args.ServerURL, "serve-url", "localhost:8080", "")
	pflag.Parse()
	return args
}
