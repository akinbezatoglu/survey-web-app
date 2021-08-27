package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config represents MongoDB configuration
type Config struct {
	ConnectionURI string `yaml:"connection_uri"`
	DatabaseName  string `yaml:"database_name"`
}

// DB represents the structure of the database
type DB struct {
	config      *Config
	client      *mongo.Client
	collections *Collections
}

// Collections represents all needed db collections
type Collections struct {
	users     *mongo.Collection
	forms     *mongo.Collection
	questions *mongo.Collection
	answers   *mongo.Collection
}

// NewConnection creates a new database connection
func NewConnection(config *Config) (*DB, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.ConnectionURI))
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	userIndexOptions := options.Index()
	userIndexOptions.SetUnique(true)

	users := client.Database(config.DatabaseName).Collection("User")
	forms := client.Database(config.DatabaseName).Collection("Form")
	questions := client.Database(config.DatabaseName).Collection("Question")
	answers := client.Database(config.DatabaseName).Collection("Answer")
	users.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: userIndexOptions,
	})

	collections := &Collections{
		users:     users,
		forms:     forms,
		questions: questions,
		answers:   answers,
	}

	return &DB{
		config:      config,
		client:      client,
		collections: collections,
	}, nil
}

// CloseConnection closes the database connection
func (db *DB) CloseConnection() error {
	err := db.client.Disconnect(context.Background())
	if err != nil {
		return err
	}

	return nil
}
