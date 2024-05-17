package main

import (
	"context"
	"encoding/json"
	"github.com/latonaio/data-platform-request-handler-kube/config"
	"github.com/latonaio/golang-logging-library-for-data-platform/logger"
	rabbitmq "github.com/latonaio/rabbitmq-golang-client-for-data-platform"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func main() {
	l := logger.NewLogger()
	conf := config.NewConf()

	rmq, err := rabbitmq.NewRabbitmqClient(
		conf.RMQ.URL(),
		conf.RMQ.QueueFrom(),
		"",
		conf.RMQ.QueueToSQL(),
		0,
	)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Close()
	iter, err := rmq.Iterator()
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rmq.Stop()

	clientOptions := options.Client().ApplyURI(conf.DB.DSN())
	mongoDbClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		l.Fatal(err.Error())
	}

	l.Info("Ready to receive message")

	for msg := range iter {
		l.Info("Received a message")

		start := time.Now()

		var data map[string]interface{}
		if err := json.Unmarshal(msg.Raw(), &data); err != nil {
			l.Error("Error parsing message: %v", err)
			continue
		}

		runtimeSessionID, ok := data["runtime_session_id"].(string)
		if !ok {
			l.Error("runtime_session_id is missing or not a string")
			continue
		}
		data["requestID"] = runtimeSessionID
		if err := insertDataToMongoDB(
			data,
			mongoDbClient,
			conf.DB.DatabaseName(),
			conf.DB.CollectionName(),
			l,
		); err != nil {
			log.Printf("Error inserting data to MongoDB: %v", err)
		}

		if err != nil {
			msg.Fail()
			continue
		}

		msg.Success()
		l.Info("process time %v\n", time.Since(start).Milliseconds())
	}
}

func insertDataToMongoDB(
	data map[string]interface{},
	client *mongo.Client,
	dbName string,
	collectionName string,
	log *logger.Logger,
) error {
	collection := client.Database(dbName).Collection(collectionName)
	_, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		return err
	}
	log.Info("Document inserted to MongoDB: %v", data["requestID"])
	return nil
}
