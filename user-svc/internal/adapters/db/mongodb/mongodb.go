package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func Adapter(uri, database string) (*MongoDB, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &MongoDB{
		client: client,
		db:     client.Database(database),
	}, nil
}

func (m *MongoDB) LogMessage(ctx context.Context, msg string) error {
	collection := m.db.Collection("logs")
	_, err := collection.InsertOne(ctx, bson.M{"message": msg, "timestamp": time.Now()})
	return err
}
