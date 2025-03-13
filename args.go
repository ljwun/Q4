package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"q4/api"
	"q4/api/openapi"
)

func ParseArgs() (*Args, error) {
	const op = "ParseArgs"

	// server config
	pflag.String("server-url", "0.0.0.0:8080", "")
	pflag.String("instance-id", "", "")

	// auth config
	_, defaultPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate default ed25519 private key: %w", op, err)
	}
	pflag.String("auth-issuer", "q4-api", "")
	pflag.String("auth-audience", "q4-ui", "")
	pflag.BytesBase64("auth-private-key", defaultPrivateKey, "")
	pflag.Duration("auth-expire-duration", 3*time.Hour, "")

	// oidc config
	oidcProviders := []openapi.SSOProvider{
		openapi.Internal,
		openapi.Google,
		// todo: Github 沒有提供 OIDC ，需要透過 OAuth2 和 Github API 來實作
		// openapi.Github,
		// todo: Microsoft 有提供 OIDC ，但 coreos-oidc 對於 Microsoft 的實作有問題
		// 參考: https://github.com/coreos/go-oidc/issues/344
		// openapi.Microsoft,
	}
	for _, provider := range oidcProviders {
		pflag.String(fmt.Sprintf("oidc-%s-issuer-url", provider), "", "")
		pflag.String(fmt.Sprintf("oidc-%s-client-id", provider), "", "")
		pflag.String(fmt.Sprintf("oidc-%s-client-secret", provider), "", "")
	}

	// s3 config
	pflag.String("s3-endpoint", "", "")
	pflag.String("s3-bucket", "", "")
	pflag.String("s3-public-base-url", "", "")
	pflag.String("s3-access-key-id", "", "")
	pflag.String("s3-secret-access-key", "", "")
	pflag.Int64("s3-rate-limit-per-hour", 3, "")

	// db config
	pflag.String("db-user", "", "")
	pflag.String("db-password", "", "")
	pflag.String("db-host", "", "")
	pflag.Int("db-port", 5432, "")
	pflag.String("db-database", "", "")
	pflag.String("db-schema", "", "")

	// redis config
	pflag.String("redis-addr", "", "")
	pflag.String("redis-password", "", "")
	pflag.Int("redis-db", 15, "")
	pflag.Duration("redis-expire-time", 3*24*time.Hour, "")
	pflag.String("redis-key-prefix", "q4:", "")
	pflag.String("redis-consumer-group", "q4-bid-group", "")

	// redis stream keys
	pflag.String("redis-stream-key-for-bid", "q4-shared-bid-stream", "")

	// bind pflag to viper
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("Q4")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// parse arguments
	// auth private key
	authPrivateKey, err := base64.StdEncoding.DecodeString(viper.GetString("auth-private-key"))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to decode auth private key: %w", op, err)
	}
	if len(authPrivateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("%s: invalid auth private key size", op)
	}
	// oidc providers
	providerConfigs := make(map[string]api.OIDCProviderConfig)
	for _, provider := range oidcProviders {
		providerConfig := api.OIDCProviderConfig{
			IssuerURL:    viper.GetString(fmt.Sprintf("oidc-%s-issuer-url", provider)),
			ClientID:     viper.GetString(fmt.Sprintf("oidc-%s-client-id", provider)),
			ClientSecret: viper.GetString(fmt.Sprintf("oidc-%s-client-secret", provider)),
		}
		if providerConfig.IssuerURL == "" || providerConfig.ClientID == "" || providerConfig.ClientSecret == "" {
			continue
		}
		providerConfigs[string(provider)] = providerConfig
	}

	// initial arguments
	return &Args{
		ServerURL: viper.GetString("server-url"),
		ServerConfig: api.ServerConfig{
			ID: viper.GetString("instance-id"),
			Auth: api.AuthConfig{
				Issuer:         viper.GetString("auth-issuer"),
				PrivateKey:     authPrivateKey,
				Audience:       viper.GetString("auth-audience"),
				ExpireDuration: viper.GetDuration("auth-expire-duration"),
			},
			OIDC: api.OIDCConfig{
				Providers: providerConfigs,
			},
			S3: api.S3Config{
				Endpoint:         viper.GetString("s3-endpoint"),
				Bucket:           viper.GetString("s3-bucket"),
				PublicBaseURL:    viper.GetString("s3-public-base-url"),
				AccessKeyID:      viper.GetString("s3-access-key-id"),
				SecretAccessKey:  viper.GetString("s3-secret-access-key"),
				RateLimitPerHour: viper.GetInt64("s3-rate-limit-per-hour"),
			},
			DB: api.DBConfig{
				User:     viper.GetString("db-user"),
				Password: viper.GetString("db-password"),
				Host:     viper.GetString("db-host"),
				Port:     viper.GetInt("db-port"),
				Database: viper.GetString("db-database"),
				Schema:   viper.GetString("db-schema"),
			},
			Redis: api.RedisConfig{
				Addr:          viper.GetString("redis-addr"),
				Password:      viper.GetString("redis-password"),
				DB:            viper.GetInt("redis-db"),
				ExpireTime:    viper.GetDuration("redis-expire-time"),
				KeyPrefix:     viper.GetString("redis-key-prefix"),
				ConsumerGroup: viper.GetString("redis-consumer-group"),
				StreamKeys: api.RedisStreamKeys{
					BidStream: viper.GetString("redis-stream-key-for-bid"),
				},
			},
		},
	}, nil
}

type Args struct {
	ServerURL    string
	ServerConfig api.ServerConfig
}

func (args Args) Validate() bool {
	// todo: validate arguments
	return true
}
