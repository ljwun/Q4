package api

import (
	"crypto/ed25519"
	"time"
)

type ServerConfig struct {
	// 用於識別不同的服務實例
	ID string

	Auth  AuthConfig
	OIDC  OIDCConfig
	S3    S3Config
	DB    DBConfig
	Redis RedisConfig
}

type AuthConfig struct {
	Issuer         string
	Audience       string
	PrivateKey     ed25519.PrivateKey
	ExpireDuration time.Duration
}

type OIDCConfig struct {
	Providers map[string]OIDCProviderConfig
}

type OIDCProviderConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
}

type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	Bucket          string
	PublicBaseURL   string

	RateLimitPerHour int64
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Schema   string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int

	ExpireTime time.Duration

	KeyPrefix     string
	ConsumerGroup string
	StreamKeys    RedisStreamKeys
}

type RedisStreamKeys struct {
	BidStream string
}
