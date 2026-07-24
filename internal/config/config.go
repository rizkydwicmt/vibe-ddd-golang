package config

import (
	"reflect"
	"strings"
	"time"

	"vibe-ddd-golang/internal/common/enum"
	"vibe-ddd-golang/internal/pkg/validation"

	"github.com/spf13/viper"
)

const (
	DefaultStartTimeout = 15 * time.Second
	DefaultStopTimeout  = 10 * time.Second
)

type Config struct {
	App       AppConfig       `mapstructure:"app" validate:"required"`
	Server    ServerConfig    `mapstructure:"api" validate:"required"`
	Database  DatabaseConfig  `mapstructure:"database" validate:"required"`
	Logger    LoggerConfig    `mapstructure:"logger" validate:"required"`
	Redis     RedisConfig     `mapstructure:"redis" validate:"required"`
	Rabbit    RabbitConfig    `mapstructure:"rabbitmq" validate:"required"`
	Migration MigrationConfig `mapstructure:"migration" validate:"required"`
}

type AppConfig struct {
	Name        string       `mapstructure:"name" validate:"required"`
	Environment enum.EnvEnum `mapstructure:"environment" validate:"enum" default:"development" normalize:"trim,lower"`
	Timezone    string       `mapstructure:"timezone" validate:"timezone" default:"UTC"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host" default:"localhost"`
	Port         int           `mapstructure:"port" validate:"gte=1,lte=65535" default:"8080"`
	GRPCPort     int           `mapstructure:"grpc_port" validate:"gte=1,lte=65535" default:"9090"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"gt=0" default:"10s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"gt=0" default:"10s"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"gt=0" default:"60s"`
}

type DatabaseConfig struct {
	Driver          DatabaseDriverEnum `mapstructure:"driver" validate:"enum" normalize:"trim,lower"`
	Host            string             `mapstructure:"host" validate:"required"`
	Port            int                `mapstructure:"port" validate:"gte=1,lte=65535" default:"5432"`
	User            string             `mapstructure:"user" validate:"required"`
	Password        string             `mapstructure:"password"`
	DBName          string             `mapstructure:"db_name" validate:"required"`
	SSLMode         string             `mapstructure:"ssl_mode" default:"disable"`
	Timezone        string             `mapstructure:"timezone" validate:"timezone" default:"UTC"`
	MaxOpenConns    int                `mapstructure:"max_open_conns" validate:"gte=0" default:"20"`
	MaxIdleConns    int                `mapstructure:"max_idle_conns" validate:"gte=0,ltefield=MaxOpenConns" default:"5"`
	ConnMaxLifetime time.Duration      `mapstructure:"conn_max_lifetime" validate:"gte=0" default:"1h"`
	ConnMaxIdleTime time.Duration      `mapstructure:"conn_max_idle_time" validate:"gte=0" default:"30m"`
	// Sync runs GORM auto-migration of the owned entities on boot (dev only).
	// Production uses the out-of-band `cmd/migration` Atlas step.
	Sync bool `mapstructure:"sync" default:"true"`
	// Cache enables the GORM query cache plugin (memory-backed by default).
	Cache     bool          `mapstructure:"cache" default:"false"`
	CacheTime time.Duration `mapstructure:"cache_time" validate:"gte=0" default:"5m"`
}

type DatabaseDriverEnum string

const (
	DatabaseDriverPostgres DatabaseDriverEnum = "postgres"
	DatabaseDriverMySQL    DatabaseDriverEnum = "mysql"
)

func (e DatabaseDriverEnum) ToString() string {
	switch e {
	case DatabaseDriverPostgres, DatabaseDriverMySQL:
		return string(e)
	default:
		return ""
	}
}

func (e DatabaseDriverEnum) IsValid() bool {
	switch e {
	case DatabaseDriverPostgres, DatabaseDriverMySQL:
		return true
	default:
		return false
	}
}

type LoggerConfig struct {
	Level      enum.LogLevelEnum  `mapstructure:"level" validate:"enum" default:"info" normalize:"trim,lower"`
	Format     enum.LogFormatEnum `mapstructure:"format" validate:"enum" default:"json" normalize:"trim,lower"`
	OutputPath string             `mapstructure:"output_path" default:"stdout"`
}

type RedisConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port" validate:"gte=1,lte=65535" default:"6379"`
	Password  string `mapstructure:"password"`
	DB        int    `mapstructure:"db" validate:"gte=0" default:"0"`
	KeyPrefix string `mapstructure:"key_prefix" default:"vibe:"`
	PoolSize  int    `mapstructure:"pool_size" validate:"gte=0" default:"10"`
}

type RabbitConfig struct {
	URI           string `mapstructure:"uri"`
	EventExchange string `mapstructure:"event_exchange" default:"vibe.events"`
}

type MigrationConfig struct {
	MigrationsDir string `mapstructure:"migrations_dir" default:"migrations"`
	Debug         bool   `mapstructure:"debug" default:"false"`
}

// IsProduction reports whether the process is running in the production env.
func (c *Config) IsProduction() bool {
	return c.App.Environment == enum.PRODUCTION
}

func (c *Config) Validate() error {
	normalizeFromTags(reflect.ValueOf(c).Elem())
	if err := validation.Validate(c); err != nil {
		return err
	}
	return nil
}

func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaultsFromTags(v, reflect.TypeOf(Config{}), "")
	if err := bindEnvFields(v, reflect.TypeOf(Config{}), ""); err != nil {
		return nil, err
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func bindEnvFields(v *viper.Viper, t reflect.Type, prefix string) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		path := field.Tag.Get("mapstructure")
		if path == "" {
			path = strings.ToLower(field.Name)
		}
		if prefix != "" {
			path = prefix + "." + path
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct && fieldType != reflect.TypeOf(time.Duration(0)) {
			if err := bindEnvFields(v, fieldType, path); err != nil {
				return err
			}
			continue
		}

		envName := strings.ToUpper(strings.ReplaceAll(path, ".", "_"))
		if err := v.BindEnv(path, envName); err != nil {
			return err
		}
	}
	return nil
}

func normalizeFromTags(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := v.Type().Field(i)

		if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(time.Duration(0)) {
			normalizeFromTags(field)
			continue
		}
		if !field.CanSet() || field.Kind() != reflect.String {
			continue
		}

		value := field.String()
		for _, op := range strings.Split(structField.Tag.Get("normalize"), ",") {
			switch strings.TrimSpace(op) {
			case "trim":
				value = strings.TrimSpace(value)
			case "lower":
				value = strings.ToLower(value)
			}
		}
		field.SetString(value)
	}
}

func setDefaultsFromTags(v *viper.Viper, t reflect.Type, prefix string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		path := field.Tag.Get("mapstructure")
		if path == "" {
			path = strings.ToLower(field.Name)
		}
		if prefix != "" {
			path = prefix + "." + path
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct && fieldType != reflect.TypeOf(time.Duration(0)) {
			setDefaultsFromTags(v, fieldType, path)
			continue
		}

		if defaultValue, ok := field.Tag.Lookup("default"); ok {
			v.SetDefault(path, defaultValue)
		}
	}
}
