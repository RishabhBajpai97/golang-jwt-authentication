package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInstance() *mongo.Client {
	err:=godotenv.Load("./.env")
	if err!=nil{
		log.Fatal(err)
	}
	MongoDb := os.Getenv("DB_URI")
	clientOptions := options.Client().ApplyURI(MongoDb)
	 client,_ := mongo.Connect(context.Background(),clientOptions)
	if err!=nil{
		log.Fatal("Error connectying to database")
	}
	fmt.Println("Connected to Database")
	return client
}
var Client = DBInstance()

func OpenCollection(client *mongo.Client, collectionName string, ) *mongo.Collection  {
	var collection *mongo.Collection = client.Database("jwt").Collection(collectionName)
	return collection
}