package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type DBConfig struct {
	Driver string
	Hosts  host
	Creds  *DBCreds
	DBName string
	Port   string
}

type host struct {
	Master   string
	Sources  []string
	Replicas []string
}

type DBCreds struct {
	Username string
	Password string
}

var DbViper *viper.Viper

func loadDbCreds() *DBCreds {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	DbViper = viper.New()

	DbViper.SetConfigType("env")
	DbViper.SetConfigName("db")
	DbViper.AddConfigPath("/vault/secrets/")
	DbViper.AddConfigPath("./")

	err := DbViper.ReadInConfig()
	if err != nil {
		sugar.Errorw("Error occurred ", "err", err)
	}
	DbViper.WatchConfig()
	DbViper.OnConfigChange(func(in fsnotify.Event) {
		logger.Info("database config changed", zap.String("name", in.Name), zap.Any("op", in.Op))
	})
	return &DBCreds{
		Username: DbViper.GetString("db_username"),
		Password: DbViper.GetString("db_password"),
	}

}

// LoadDatabaseConfig returns db configs
func LoadDatabaseConfig() *DBConfig {
	dbcreds := loadDbCreds()
	var dbconfig *DBConfig

	if CFG.V.GetString("env") == "hosted" {
		dbconfig = &DBConfig{
			Driver: CFG.V.GetString("database.driver"),
			Hosts: host{
				Master:   CFG.V.GetString("database.hosts.master"),
				Sources:  CFG.V.GetStringSlice("database.hosts.sources"),
				Replicas: CFG.V.GetStringSlice("database.hosts.replicas"),
			},
			Creds:  dbcreds,
			DBName: CFG.V.GetString("database.dbname"),
			Port:   CFG.V.GetString("database.port"),
		}

	} else {
		dbconfig = &DBConfig{
			Driver: CFG.V.GetString("database.driver"),
			Hosts: host{
				Master: CFG.V.GetString("database.host"),
			},
			Creds:  dbcreds,
			DBName: CFG.V.GetString("database.dbname"),
			Port:   CFG.V.GetString("database.port"),
		}
	}
	return dbconfig

}

// LoadDatabaseVaultConfig loads database config and secrets from a single vault kv store
func LoadDatabaseVaultConfig() *DBConfig {
	dbconfig := &DBConfig{
		Driver: CFG.V.GetString("DATABASE_DRIVER"),
		Hosts: host{
			Master:   CFG.V.GetString("DATABASE_SOURCE"),
			Sources:  []string{CFG.V.GetString("DATABASE_SOURCE")},
			Replicas: []string{CFG.V.GetString("DATABASE_REPLICA")},
		},
		Creds: &DBCreds{
			Username: CFG.V.GetString("DB_USERNAME"),
			Password: CFG.V.GetString("DB_PASSWORD"),
		},
		DBName: CFG.V.GetString("DB_NAME"),
		Port:   CFG.V.GetString("DB_PORT"),
	}

	return dbconfig
}
