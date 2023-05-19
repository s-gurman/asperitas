package post

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoCursor interface {
	All(context.Context, interface{}) error
}

type MongoSingleResult interface {
	Decode(v interface{}) error
	Err() error
}

type MongoCollection interface {
	Find(context.Context, interface{}) (MongoCursor, error)
	FindOne(context.Context, interface{}) MongoSingleResult
	InsertOne(context.Context, interface{}) (interface{}, error)
	UpdateOne(context.Context, interface{}, interface{}) (interface{}, error)
	DeleteOne(context.Context, interface{}) (interface{}, error)
}

type mongoCursor struct {
	cs *mongo.Cursor
}

type mongoSingleResult struct {
	sr *mongo.SingleResult
}

type mongoCollection struct {
	cln *mongo.Collection
}

func newMongoCollection(coll *mongo.Collection) MongoCollection {
	return &mongoCollection{cln: coll}
}

func (mcs *mongoCursor) All(ctx context.Context, v interface{}) error {
	return mcs.cs.All(ctx, v)
}

func (msr *mongoSingleResult) Decode(v interface{}) error {
	return msr.sr.Decode(v)
}

func (msr *mongoSingleResult) Err() error {
	return msr.sr.Err()
}

func (mc *mongoCollection) Find(ctx context.Context, filter interface{}) (MongoCursor, error) {
	cursor, err := mc.cln.Find(ctx, filter)
	return &mongoCursor{cs: cursor}, err
}

func (mc *mongoCollection) FindOne(ctx context.Context, filter interface{}) MongoSingleResult {
	singleResult := mc.cln.FindOne(ctx, filter)
	return &mongoSingleResult{sr: singleResult}
}

func (mc *mongoCollection) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	insertResult, err := mc.cln.InsertOne(ctx, document)
	return insertResult, err
}

func (mc *mongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (interface{}, error) {
	updateResult, err := mc.cln.UpdateOne(ctx, filter, update)
	return updateResult, err
}

func (mc *mongoCollection) DeleteOne(ctx context.Context, filter interface{}) (interface{}, error) {
	deleteResult, err := mc.cln.DeleteOne(ctx, filter)
	return deleteResult, err
}
