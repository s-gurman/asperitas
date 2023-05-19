package post

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"asperitas/internal/errs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var emptyCtx = context.Background()

type PostRepositoryMongo struct {
	coll MongoCollection
}

func NewRepoMongo(addr string) (*PostRepositoryMongo, error) {
	mongoConn, err := mongo.Connect(emptyCtx, options.Client().ApplyURI(addr))
	if err != nil {
		return nil, fmt.Errorf("mongo connect err: %w", err)
	}
	mongoColl := mongoConn.Database("vk-go").Collection("posts")
	mongoCollAbstract := newMongoCollection(mongoColl)
	return &PostRepositoryMongo{coll: mongoCollAbstract}, nil
}

func findPosts(coll MongoCollection, filter primitive.M) ([]*Post, error) {
	cursor, err := coll.Find(emptyCtx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongo find err: %w", err)
	}
	posts := []*Post{}
	if err = cursor.All(emptyCtx, &posts); err != nil {
		return nil, fmt.Errorf("mongo all err: %w", err)
	}
	return posts, nil
}

func findPost(coll MongoCollection, id string) (*Post, error) {
	res := coll.FindOne(emptyCtx, bson.M{"id": id})
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, errs.MsgError{Msg: "post not found", Status: 404}
	}
	post := &Post{}
	if err := res.Decode(post); err != nil {
		return nil, fmt.Errorf("mongo decode err: %w", err)
	}
	return post, nil
}

func (repo *PostRepositoryMongo) GetAll() ([]*Post, error) {
	posts, err := findPosts(repo.coll, bson.M{})
	if err != nil {
		return nil, err
	}
	sort.Slice(posts, func(i, j int) bool {
		if posts[i].Score == posts[j].Score {
			return posts[i].Created.Before(posts[j].Created)
		}
		return posts[i].Score > posts[j].Score
	})
	return posts, nil
}

func (repo *PostRepositoryMongo) AddPost(p *Post) error {
	if _, err := repo.coll.InsertOne(emptyCtx, p); err != nil {
		return fmt.Errorf("mongo insert one err: %w", err)
	}
	return nil
}

func (repo *PostRepositoryMongo) GetByCategory(categoryName string) ([]*Post, error) {
	posts, err := findPosts(repo.coll, bson.M{"category": categoryName})
	if err != nil {
		return nil, err
	}
	sort.Slice(posts, func(i, j int) bool {
		if posts[i].Score == posts[j].Score {
			return posts[i].Created.Before(posts[j].Created)
		}
		return posts[i].Score > posts[j].Score
	})
	return posts, nil
}

func (repo *PostRepositoryMongo) GetByID(id string) (*Post, error) {
	p, err := findPost(repo.coll, id)
	if err != nil {
		return nil, err
	}
	p.Views++
	if _, err := repo.coll.UpdateOne(
		emptyCtx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{"views": p.Views}},
	); err != nil {
		return nil, fmt.Errorf("mongo update one err: %w", err)
	}
	return p, nil
}

func (repo *PostRepositoryMongo) DeletePost(postID, userID string) error {
	p, err := findPost(repo.coll, postID)
	if err != nil {
		return err
	}
	if p.Author.ID != userID {
		return errs.MsgError{Msg: "unauthorized", Status: 401}
	}
	if _, err := repo.coll.DeleteOne(emptyCtx, bson.M{"id": postID}); err != nil {
		return fmt.Errorf("mongo delete one err: %w", err)
	}
	return errs.MsgError{Msg: "success", Status: 200}
}

func (repo *PostRepositoryMongo) AddComment(postID string, comm *Comment) (*Post, error) {
	p, err := findPost(repo.coll, postID)
	if err != nil {
		return nil, err
	}
	if err := p.Comments.Add(comm); err != nil {
		return nil, err
	}
	if _, err := repo.coll.UpdateOne(
		emptyCtx,
		bson.M{"id": postID},
		bson.M{"$set": bson.M{"comments": p.Comments}},
	); err != nil {
		return nil, fmt.Errorf("mongo update one err: %w", err)
	}
	return p, nil
}

func (repo *PostRepositoryMongo) DeleteComment(postID, commID, userID string) (*Post, error) {
	p, err := findPost(repo.coll, postID)
	if err != nil {
		return nil, err
	}
	if err := p.Comments.Delete(commID, userID); err != nil {
		return nil, err
	}
	if _, err := repo.coll.UpdateOne(
		emptyCtx,
		bson.M{"id": postID},
		bson.M{"$set": bson.M{"comments": p.Comments}},
	); err != nil {
		return nil, fmt.Errorf("mongo update one err: %w", err)
	}
	return p, nil
}

func (repo *PostRepositoryMongo) UpvotePost(postID, userID string) (*Post, error) {
	p, err := findPost(repo.coll, postID)
	if err != nil {
		return nil, err
	}
	if err := p.Votes.Upvote(userID); err != nil {
		return nil, err
	}
	p.updatePostScore()
	if _, err := repo.coll.UpdateOne(
		emptyCtx,
		bson.M{"id": postID},
		bson.M{"$set": bson.M{
			"votes":        p.Votes,
			"score":        p.Score,
			"likespercent": p.LikesPercent,
		}},
	); err != nil {
		return nil, fmt.Errorf("mongo update one err: %w", err)
	}
	return p, nil
}

func (repo *PostRepositoryMongo) DownvotePost(postID, userID string) (*Post, error) {
	p, err := findPost(repo.coll, postID)
	if err != nil {
		return nil, err
	}
	if err := p.Votes.Downvote(userID); err != nil {
		return nil, err
	}
	p.updatePostScore()
	if _, err := repo.coll.UpdateOne(
		emptyCtx,
		bson.M{"id": postID},
		bson.M{"$set": bson.M{
			"votes":        p.Votes,
			"score":        p.Score,
			"likespercent": p.LikesPercent,
		}},
	); err != nil {
		return nil, fmt.Errorf("mongo update one err: %w", err)
	}
	return p, nil
}

func (repo *PostRepositoryMongo) UnvotePost(postID, userID string) (*Post, error) {
	p, err := findPost(repo.coll, postID)
	if err != nil {
		return nil, err
	}
	if err := p.Votes.Unvote(userID); err != nil {
		return nil, err
	}
	p.updatePostScore()
	if _, err := repo.coll.UpdateOne(
		emptyCtx,
		bson.M{"id": postID},
		bson.M{"$set": bson.M{
			"votes":        p.Votes,
			"score":        p.Score,
			"likespercent": p.LikesPercent,
		}},
	); err != nil {
		return nil, fmt.Errorf("mongo update one err: %w", err)
	}
	return p, nil
}

func (repo *PostRepositoryMongo) GetByUser(username string) ([]*Post, error) {
	posts, err := findPosts(repo.coll, bson.M{"author.username": username})
	if err != nil {
		return nil, err
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Created.After(posts[j].Created)
	})
	return posts, nil
}
