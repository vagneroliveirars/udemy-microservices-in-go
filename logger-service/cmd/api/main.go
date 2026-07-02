package main

import (
	"log"
	"logger-service/data"

	"go.mongodb.org/mongo-driver/mongo"
)

const {
	webPort = "80"
	rpcPort = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
}

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// Connect to MongoDB
	mongoClient, err := connectToMongoDB()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// create a context in order to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	} ()

	app := Config{
		Models: data.New(client),
	}

	// start web server
	go app.serve()
}

func (app *Config) serve() {
	log.Println("Starting broker service on port", webPort)

	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectToMongoDB() (*mongo.Client, error) {
	// Create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.setAuth(options.Credential{
		Username: "admin",
		Password: "password"
	})

	// Connect to MongoDB
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Failed to connect to MongoDB:", err)
		return nil, err
	}

	return c, nil
}