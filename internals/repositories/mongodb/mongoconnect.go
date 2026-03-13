package mongodb

import (
	"context"
	"log"

	"github.com/Sandwichzzy/school_manager_system_grpc/pkg/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateMongoClient() (*mongo.Client, error) {
	ctx := context.Background()
	//client, err := mongo.Connect(ctx, options.Client().ApplyURI("username:password@mongodb://localhost:27017"))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to MongoDB:")
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error pinging MongoDB:")
	}
	log.Println("Successfully connected to MongoDB")
	return client, nil
}
