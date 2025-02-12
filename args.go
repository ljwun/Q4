package main

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"q4/api"
)

func ParseArgs() Args {
	// server config
	pflag.String("server-url", "0.0.0.0:8080", "")

	// oidc config
	pflag.String("oidc-issuer-url", "", "")
	pflag.String("oidc-client-id", "", "")
	pflag.String("oidc-client-secret", "", "")

	// s3 config
	pflag.String("s3-endpoint", "", "")
	pflag.String("s3-bucket", "", "")
	pflag.String("s3-public-base-url", "", "")
	pflag.String("s3-access-key-id", "", "")
	pflag.String("s3-secret-access-key", "", "")

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

	// redis stream keys
	pflag.String("redis-stream-key-for-sse", "q4-shared-sse-stream", "")

	// bind pflag to viper
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("Q4")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// initial arguments
	return Args{
		ServerURL: viper.GetString("server-url"),
		ServerConfig: api.ServerConfig{
			OIDC: api.OIDCConfig{
				IssuerURL:    viper.GetString("oidc-issuer-url"),
				ClientID:     viper.GetString("oidc-client-id"),
				ClientSecret: viper.GetString("oidc-client-secret"),
			},
			S3: api.S3Config{
				Endpoint:        viper.GetString("s3-endpoint"),
				Bucket:          viper.GetString("s3-bucket"),
				PublicBaseURL:   viper.GetString("s3-public-base-url"),
				AccessKeyID:     viper.GetString("s3-access-key-id"),
				SecretAccessKey: viper.GetString("s3-secret-access-key"),
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
				Addr:     viper.GetString("redis-addr"),
				Password: viper.GetString("redis-password"),
				DB:       viper.GetInt("redis-db"),
				StreamKeys: api.RedisStreamKeys{
					SSE: viper.GetString("redis-stream-key-for-sse"),
				},
			},
		},
	}
}

type Args struct {
	ServerURL    string
	ServerConfig api.ServerConfig
}

func (args Args) Validate() bool {
	return args.ServerURL != "" && args.ServerConfig.OIDC.IssuerURL != "" && args.ServerConfig.OIDC.ClientID != "" && args.ServerConfig.OIDC.ClientSecret != ""
}
