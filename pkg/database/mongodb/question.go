package mongodb

import (
	"context"

	"github.com/akinbezatoglu/survey-web-app/pkg/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateUser creates a new user
func (db *DB) CreateQuestion(q *model.Question) (string, error) {
	//u.ID = primitive.NewObjectID().Hex()

	result, err := db.collections.questions.InsertOne(context.Background(), q)
	if err != nil {
		return "", err
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// GetUser returns a user
func (db *DB) GetQuestion(id string) (*model.Question, error) {
	var ques model.Question
	docID, _ := primitive.ObjectIDFromHex(id)
	cursor := db.collections.questions.FindOne(
		context.Background(),
		bson.D{primitive.E{
			Key:   "_id",
			Value: docID,
		}},
	)
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}
	err := cursor.Decode(&ques)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &ques, nil
}

// SaveUser saves the given user struct
func (db *DB) SaveQuestion(q *model.Question) error {
	cursor := db.collections.questions.FindOneAndReplace(
		context.Background(),
		bson.D{primitive.E{
			Key:   "_id",
			Value: q.ID,
		}},
		q,
	)
	if cursor.Err() != nil {
		return cursor.Err()
	}
	return nil
}

// DeleteUser deletes the user with the given id
func (db *DB) DeleteQuestion(id primitive.ObjectID) error {
	cursor := db.collections.questions.FindOneAndDelete(
		context.Background(),
		bson.D{primitive.E{
			Key:   "_id",
			Value: id,
		}},
	)
	if cursor.Err() != nil {
		return cursor.Err()
	}
	return nil
}
