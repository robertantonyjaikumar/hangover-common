package database

import (
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/robertantonyjaikumar/hangover-common/config"
	"github.com/robertantonyjaikumar/hangover-common/logger"
	"go.uber.org/zap"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	gormtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorm.io/gorm.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"moul.io/zapgorm2"
)

var (
	Db     = InitDb()
	LOCAL  = "local"
	HOSTED = "hosted"
)

func InitDb() *gorm.DB {
	var dbConfig *config.DBConfig
	if config.CFG.V.GetBool("database.single_source") {
		dbConfig = config.LoadDatabaseVaultConfig()
	} else {
		dbConfig = config.LoadDatabaseConfig()

	}

	if config.CFG.V.GetString("env") == HOSTED {
		return connectMultipleDB(dbConfig)
	} else {
		return connectDB(dbConfig)
	}
}

// Returns an initialized *gorm.DB struct
func connectDB(database *config.DBConfig) *gorm.DB {

	dsn := database.Driver + "://" + database.Creds.Username + ":" + database.Creds.Password + "@" + database.Hosts.Master + ":" + database.Port + "/" + database.DBName + "?" + "application_name=" + config.CFG.GetServiceName()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Error connecting to database: connection url error", zap.Error(err))
		return nil
	}
	return db
}

// Migrations Create a migration struct object
type Migrations struct {
	DB     *gorm.DB
	Models []interface{}
}

// RunMigrations runs migrations
func RunMigrations(migrations Migrations) {
	for _, model := range migrations.Models {
		err := migrations.DB.AutoMigrate(model)
		if err != nil {
			logger.Error("Could not migrate %s", zap.Error(err))

		}
	}
}

func connectMultipleDB(database *config.DBConfig) *gorm.DB {
	gormLogger := zapgorm2.New(logger.GetZapLogger())
	gormLogger.SetAsDefault()
	dsn := database.Driver + "://" + database.Creds.Username + ":" + database.Creds.Password + "@" + database.Hosts.Master + ":" + database.Port + "/" + database.DBName + "?" + "application_name=" + config.CFG.GetServiceName()
	// Register augments the provided driver with tracing, enabling it to be loaded by
	// gormtrace.Open.
	sqltrace.Register(
		"pgx",
		&stdlib.Driver{},
		sqltrace.WithServiceName(config.CFG.GetServiceName()),
	)
	sqlDb, err := sqltrace.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("Error occurred", zap.Error(err))
	}
	db, err := gormtrace.Open(
		postgres.New(postgres.Config{Conn: sqlDb}),
		&gorm.Config{Logger: gormLogger},
	)
	if err != nil {
		logger.Error("Error connecting to database: connection url error ", zap.Error(err))
		return nil
	}

	var (
		replicas, sources []gorm.Dialector
	)

	//Create db sources(write instances) from config
	for _, host := range database.Hosts.Sources {
		dsn := database.Driver + "://" + database.Creds.Username + ":" + database.Creds.Password + "@" + host + ":" + database.Port + "/" + database.DBName + "?" + "application_name=" + config.CFG.GetServiceName()
		sources = append(sources, postgres.Open(dsn))
	}

	//Create db replicas(read instances) from config
	for _, host := range database.Hosts.Replicas {
		dsn := database.Driver + "://" + database.Creds.Username + ":" + database.Creds.Password + "@" + host + ":" + database.Port + "/" + database.DBName + "?" + "application_name=" + config.CFG.GetServiceName()
		replicas = append(replicas, postgres.Open(dsn))
	}
	//logger.Info("DB URLs", zap.Any("sources", sources), zap.Any("replicas", replicas))
	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  sources,
		Replicas: replicas,
		// sources/replicas load balancing policy
		Policy: dbresolver.RandomPolicy{},
	}))
	if err != nil {
		logger.Error("Error connecting to database ", zap.Error(err))
		return nil
	}

	return db
}
