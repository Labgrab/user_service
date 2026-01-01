package config

type Config struct {
	Port   int    `env:"PORT,required"`
	DBConn string `env:"DB_CONNECT"`
}
