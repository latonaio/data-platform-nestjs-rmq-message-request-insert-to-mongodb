package config

import (
	"fmt"
	"os"
)

type Database struct {
	dbName         string
	collectionName string
	address        string
	port           string
}

func newDatabase() *Database {
	return &Database{
		dbName:         os.Getenv("MONGO_DB_NAME"),
		collectionName: os.Getenv("MONGO_DB_COLLECTION_NAME"),
		address:        os.Getenv("MONGO_DB_ADDRESS"),
		port:           os.Getenv("MONGO_DB_PORT"),
	}
}

func (c Database) DSN() string {
	return fmt.Sprintf(
		"mongodb://%s:%s",
		c.address,
		c.port,
	)
}

func (c Database) DatabaseName() string {
	return c.dbName
}

func (c Database) CollectionName() string {
	return c.collectionName
}
