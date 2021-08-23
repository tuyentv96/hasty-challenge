package config

type Config struct {
	SQLConfig
	HTTPConfig
	RedisConfig
	LoggerConfig
	JobConfig
}

type HTTPConfig struct {
	HTTPPort   int  `envconfig:"HTTP_PORT" default:"3000"`
	HTTPLogger bool `envconfig:"HTTP_LOGGER" default:"true"`
}

type SQLConfig struct {
	SQLName     string `envconfig:"SQL_NAME" required:"true"`
	SQLAddress  string `envconfig:"SQL_ADDRESS" required:"true"`
	SQLUser     string `envconfig:"SQL_USER" required:"true"`
	SQLPassword string `envconfig:"SQL_PASSWORD" required:"true"`
}

type RedisConfig struct {
	RedisAddress        string `envconfig:"REDIS_ADDRESS" default:"localhost:6379"`
	RedisPollIntervalMs int    `envconfig:"REDIS_POLL_INTERVAL" default:"1000"`
}

type LoggerConfig struct {
	Level string `envconfig:"LOGGER_LEVEL" default:"info"`
}

type JobConfig struct {
	JobPrefetch      int64 `envconfig:"JOB_PREFETCH" default:"10"`
	TimeoutInSeconds int   `envconfig:"JOB_TIMEOUT" default:"20"`
}
