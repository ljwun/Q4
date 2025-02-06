package api

type ServerConfig struct {
	OIDC OIDCConfig
	S3   S3Config
	DB   DBConfig
}

type OIDCConfig struct {
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
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Schema   string
}
