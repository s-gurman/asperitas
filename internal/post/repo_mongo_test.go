package post

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"asperitas/internal/errs"
	"asperitas/internal/user"
	"asperitas/pkg/rand"

	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var (
	usr1   = user.User{Username: "admin1", ID: "id_admin1", Password: "passw"}
	usr2   = user.User{Username: "admin2", ID: "id_admin2", Password: "passw"}
	randID = rand.GetRandID()
)

func getDoc(v interface{}) (doc bson.D) {
	data, _ := bson.Marshal(v) // nolint:errcheck
	bson.Unmarshal(data, &doc) // nolint:errcheck
	return doc
}

func copyPost(src *Post) Post {
	votesList := make([]Vote, len(src.Votes.List))
	copy(votesList, src.Votes.List)
	commList := make(CommentList, len(src.Comments))
	copy(commList, src.Comments)
	dst := *src
	dst.Votes.List = votesList
	dst.Comments = commList
	return dst
}

func getMockService(t *testing.T) (*PostRepositoryMongo, *MockMongoCollection, *MockMongoSingleResult) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cln := NewMockMongoCollection(ctrl)
	sr := NewMockMongoSingleResult(ctrl)
	return &PostRepositoryMongo{coll: cln}, cln, sr
}

func TestGetAll_OK(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run(t.Name(), func(mt *mtest.T) {
		expect := []*Post{NewPost(usr1), NewPost(usr2), NewPost(usr1)}
		expect[0].Score = 2

		first := mtest.CreateCursorResponse(1, "db.mock", mtest.FirstBatch, getDoc(expect[0]))
		responses := []primitive.D{first}
		for i := 1; i < len(expect); i++ {
			next := mtest.CreateCursorResponse(1, "db.mock", mtest.NextBatch, getDoc(expect[i]))
			responses = append(responses, next)
		}
		last := mtest.CreateCursorResponse(0, "db.mock", mtest.NextBatch)
		responses = append(responses, last)

		mt.AddMockResponses(responses...)
		repo := &PostRepositoryMongo{coll: newMongoCollection(mt.Coll)}

		result, err := repo.GetAll()

		if err != nil {
			mt.Errorf("unexpected err: %s", err)
		}
		if len(result) != len(expect) {
			mt.Errorf("bad result len:\nwant:\t%v\nhave\t%v", len(expect), len(result))
			return
		}
		for i := 0; i < len(expect); i++ {
			expect[i].Created = result[i].Created
			if !reflect.DeepEqual(expect[i], result[i]) {
				mt.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect[i], result[i])
			}
		}
	})
}

func TestGetAll_FindErr(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run(t.Name(), func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Message: "some error",
		}))
		repo := &PostRepositoryMongo{coll: newMongoCollection(mt.Coll)}

		expect := "mongo find err"
		result, err := repo.GetAll()

		if result != nil {
			mt.Errorf("unexpected result: %#v", result)
		}
		if err == nil || !strings.HasPrefix(err.Error(), expect) {
			mt.Errorf("unexpected err:\nwant:\t%s\nhave\t%#v", expect, err)
		}
	})
}

func TestGetAll_CursorErr(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run(t.Name(), func(mt *mtest.T) {
		first := mtest.CreateCursorResponse(1, "db.mock", mtest.FirstBatch, bson.D{})

		mt.AddMockResponses(first)
		repo := &PostRepositoryMongo{coll: newMongoCollection(mt.Coll)}

		expect := "mongo all err"
		result, err := repo.GetAll()

		if result != nil {
			mt.Errorf("unexpected result: %#v", result)
		}
		if err == nil || !strings.HasPrefix(err.Error(), expect) {
			mt.Errorf("unexpected err:\nwant:\t%s\nhave\t%#v", expect, err)
		}
	})
}

func TestAddPost_OK(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run(t.Name(), func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		repo := &PostRepositoryMongo{coll: newMongoCollection(mt.Coll)}

		expect := error(nil)
		err := repo.AddPost(NewPost(user.User{}))

		if err != expect {
			mt.Errorf("unexpected err: %s", err)
		}
	})
}

func TestAddPost_InsertErr(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run(t.Name(), func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Message: "some error",
		}))
		repo := &PostRepositoryMongo{coll: newMongoCollection(mt.Coll)}

		expect := "mongo insert one err"
		err := repo.AddPost(NewPost(user.User{}))

		if err == nil || !strings.HasPrefix(err.Error(), expect) {
			mt.Errorf("unexpected err:\nwant:\t%s\nhave\t%#v", expect, err)
		}
	})
}

func TestGetByCategory_OK(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run(t.Name(), func(mt *mtest.T) {
		expect := []*Post{NewPost(usr1), NewPost(usr2), NewPost(usr1)}
		expect[0].Score = 2
		for _, post := range expect {
			post.Category = PostCategory("music")
		}

		first := mtest.CreateCursorResponse(1, "db.mock", mtest.FirstBatch, getDoc(expect[0]))
		responses := []primitive.D{first}
		for i := 1; i < len(expect); i++ {
			next := mtest.CreateCursorResponse(1, "db.mock", mtest.NextBatch, getDoc(expect[i]))
			responses = append(responses, next)
		}
		last := mtest.CreateCursorResponse(0, "db.mock", mtest.NextBatch)
		responses = append(responses, last)

		mt.AddMockResponses(responses...)
		repo := &PostRepositoryMongo{coll: newMongoCollection(mt.Coll)}

		result, err := repo.GetByCategory("music")

		if err != nil {
			mt.Errorf("unexpected err: %s", err)
		}
		if len(result) != len(expect) {
			mt.Errorf("bad result len:\nwant:\t%v\nhave\t%v", len(expect), len(result))
			return
		}
		for i := 0; i < len(expect); i++ {
			expect[i].Created = result[i].Created
			if !reflect.DeepEqual(expect[i], result[i]) {
				mt.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect[i], result[i])
			}
		}
	})
}

func TestGetByCategory_FindErr(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run(t.Name(), func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Message: "some error",
		}))
		repo := &PostRepositoryMongo{coll: newMongoCollection(mt.Coll)}

		expect := "mongo find err"
		result, err := repo.GetByCategory("music")

		if result != nil {
			mt.Errorf("unexpected result: %#v", result)
		}
		if err == nil || !strings.HasPrefix(err.Error(), expect) {
			mt.Errorf("unexpected err:\nwant:\t%s\nhave\t%#v", expect, err)
		}
	})
}

func TestGetByID_OK(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": expect.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *expect).
		Return(nil)
	expect.Views++
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": expect.ID},
			bson.M{"$set": bson.M{"views": expect.Views}},
		).
		Return(gomock.Any(), nil)

	result, err := service.GetByID(expect.ID)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, result)
	}
}

func TestGetByID_ErrNoPost(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(mongo.ErrNoDocuments)

	result, err := service.GetByID(randID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestGetByID_DecodeErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some error")

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).
		Return(expect)

	result, err := service.GetByID(randID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestGetByID_UpdateErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some error")
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)
	post.Views++
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": post.ID},
			bson.M{"$set": bson.M{"views": post.Views}},
		).
		Return(nil, expect)

	result, err := service.GetByID(post.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeletePost_OK(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "success", Status: 200}
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)
	coll.EXPECT().
		DeleteOne(emptyCtx, bson.M{"id": post.ID}).
		Return(gomock.Any(), nil)

	err := service.DeletePost(post.ID, usr1.ID)

	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeletePost_ErrNoPost(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(mongo.ErrNoDocuments)

	err := service.DeletePost(randID, usr1.ID)

	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeletePost_AuthErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	err := service.DeletePost(post.ID, usr2.ID)

	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeletePost_DeleteErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some error")
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)
	coll.EXPECT().
		DeleteOne(emptyCtx, bson.M{"id": post.ID}).
		Return(nil, expect)

	err := service.DeletePost(post.ID, usr1.ID)

	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestAddComment_OK(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := NewPost(usr1)
	comm := NewComment(usr2, "some text")
	tmp := copyPost(expect)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": expect.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	expect.Comments.Add(comm) // nolint:errcheck
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": expect.ID},
			bson.M{"$set": bson.M{"comments": expect.Comments}},
		).
		Return(gomock.Any(), nil)

	result, err := service.AddComment(expect.ID, comm)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, result)
	}
}

func TestAddComment_ErrNoPost(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(mongo.ErrNoDocuments)

	result, err := service.AddComment(randID, nil)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestAddComment_AddErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("nil comment list")
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	post.Comments = nil
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	result, err := service.AddComment(post.ID, nil)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestAddComment_UpdateErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some err")
	post := NewPost(usr1)
	comm := NewComment(usr2, "some text")
	tmp := copyPost(post)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	post.Comments.Add(comm) // nolint:errcheck
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": post.ID},
			bson.M{"$set": bson.M{"comments": post.Comments}},
		).
		Return(nil, expect)

	result, err := service.AddComment(post.ID, comm)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeleteComment_OK(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := NewPost(usr1)
	comm := NewComment(usr2, "some text")
	expect.Comments.Add(comm) // nolint:errcheck
	tmp := copyPost(expect)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": expect.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	expect.Comments.Delete(comm.ID, usr2.ID) // nolint:errcheck
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": expect.ID},
			bson.M{"$set": bson.M{"comments": expect.Comments}},
		).
		Return(gomock.Any(), nil)

	result, err := service.DeleteComment(expect.ID, comm.ID, usr2.ID)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, result)
	}
}

func TestDeleteComment_ErrNoPost(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(mongo.ErrNoDocuments)

	result, err := service.DeleteComment(randID, randID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeleteComment_DeleteErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("nil comment list")
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	post.Comments = nil
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	result, err := service.DeleteComment(post.ID, randID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeleteComment_ErrNoComment(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "comment not found", Status: 404}
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	result, err := service.DeleteComment(post.ID, randID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeleteComment_AuthErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	post := NewPost(usr1)
	comm := NewComment(usr2, "some text")
	post.Comments.Add(comm) // nolint:errcheck

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	result, err := service.DeleteComment(post.ID, comm.ID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDeleteComment_UpdateErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some err")
	post := NewPost(usr1)
	comm := NewComment(usr2, "some text")
	post.Comments.Add(comm) // nolint:errcheck

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)
	post.Comments.Delete(comm.ID, usr2.ID) // nolint:errcheck
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": post.ID},
			bson.M{"$set": bson.M{"comments": post.Comments}},
		).
		Return(nil, expect)

	result, err := service.DeleteComment(post.ID, comm.ID, usr2.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestUpvotePost_OK(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := NewPost(usr1)
	tmp := copyPost(expect)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": expect.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	expect.Votes.Upvote(usr2.ID) // nolint:errcheck
	expect.updatePostScore()
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": expect.ID},
			bson.M{"$set": bson.M{
				"votes":        expect.Votes,
				"score":        expect.Score,
				"likespercent": expect.LikesPercent,
			}},
		).
		Return(gomock.Any(), nil)

	result, err := service.UpvotePost(expect.ID, usr2.ID)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, result)
	}
}

func TestUpvotePost_ErrNoPost(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(mongo.ErrNoDocuments)

	result, err := service.UpvotePost(randID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestUpvotePost_UpvoteErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("nil vote list")
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	post.Votes.List = nil
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	result, err := service.UpvotePost(post.ID, usr2.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestUpvotePost_UpdateErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some err")
	post := NewPost(usr1)
	tmp := copyPost(post)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	post.Votes.Upvote(usr1.ID) // nolint:errcheck
	post.updatePostScore()
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": post.ID},
			bson.M{"$set": bson.M{
				"votes":        post.Votes,
				"score":        post.Score,
				"likespercent": post.LikesPercent,
			}},
		).
		Return(nil, expect)

	result, err := service.UpvotePost(post.ID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDownvotePost_OK(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := NewPost(usr1)
	tmp := copyPost(expect)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": expect.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	expect.Votes.Downvote(usr1.ID) // nolint:errcheck
	expect.updatePostScore()
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": expect.ID},
			bson.M{"$set": bson.M{
				"votes":        expect.Votes,
				"score":        expect.Score,
				"likespercent": expect.LikesPercent,
			}},
		).
		Return(gomock.Any(), nil)

	result, err := service.DownvotePost(expect.ID, usr1.ID)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, result)
	}
}

func TestDownvotePost_ErrNoPost(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(mongo.ErrNoDocuments)

	result, err := service.DownvotePost(randID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDownvotePost_DownvoteErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("nil vote list")
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	post.Votes.List = nil
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	result, err := service.DownvotePost(post.ID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestDownvotePost_UpdateErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some err")
	post := NewPost(usr1)
	tmp := copyPost(post)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	post.Votes.Downvote(usr2.ID) // nolint:errcheck
	post.updatePostScore()
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": post.ID},
			bson.M{"$set": bson.M{
				"votes":        post.Votes,
				"score":        post.Score,
				"likespercent": post.LikesPercent,
			}},
		).
		Return(nil, expect)

	result, err := service.DownvotePost(post.ID, usr2.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestUnvotePost_OK(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := NewPost(usr1)
	tmp := copyPost(expect)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": expect.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	expect.Votes.Unvote(usr1.ID) // nolint:errcheck
	expect.updatePostScore()
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": expect.ID},
			bson.M{"$set": bson.M{
				"votes":        expect.Votes,
				"score":        expect.Score,
				"likespercent": expect.LikesPercent,
			}},
		).
		Return(gomock.Any(), nil)

	result, err := service.UnvotePost(expect.ID, usr1.ID)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, result)
	}
}

func TestUnvotePost_ErrNoPost(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": randID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(mongo.ErrNoDocuments)

	result, err := service.UnvotePost(randID, usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestUnvotePost_UnvoteErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("nil vote list")
	post := NewPost(usr1)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	post.Votes.List = nil
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, *post).
		Return(nil)

	result, err := service.UnvotePost(post.ID, usr2.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !reflect.DeepEqual(expect, err) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestUnvotePost_UpdateErr(t *testing.T) {
	service, coll, sr := getMockService(t)

	expect := fmt.Errorf("some err")
	post := NewPost(usr1)
	tmp := copyPost(post)

	coll.EXPECT().
		FindOne(emptyCtx, bson.M{"id": post.ID}).
		Return(sr)
	sr.EXPECT().
		Err().
		Return(nil)
	sr.EXPECT().
		Decode(&Post{}).SetArg(0, tmp).
		Return(nil)
	post.Votes.Unvote(usr2.ID) // nolint:errcheck
	post.updatePostScore()
	coll.EXPECT().
		UpdateOne(
			emptyCtx,
			bson.M{"id": post.ID},
			bson.M{"$set": bson.M{
				"votes":        post.Votes,
				"score":        post.Score,
				"likespercent": post.LikesPercent,
			}},
		).
		Return(nil, expect)

	result, err := service.UnvotePost(post.ID, usr2.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}

func TestGetByUser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cln := NewMockMongoCollection(ctrl)
	cs := NewMockMongoCursor(ctrl)

	service := &PostRepositoryMongo{coll: cln}

	postFirst := NewPost(usr1)
	postSecond := NewPost(usr1)
	postLast := NewPost(usr1)
	posts := []*Post{postSecond, postLast, postFirst}
	expect := []*Post{postLast, postSecond, postFirst}

	cln.EXPECT().
		Find(emptyCtx, bson.M{"author.username": usr1.ID}).
		Return(cs, nil)
	cs.EXPECT().
		All(emptyCtx, &[]*Post{}).SetArg(1, posts).
		Return(nil)

	result, err := service.GetByUser(usr1.ID)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if len(result) != len(expect) {
		t.Errorf("bad result len:\nwant:\t%v\nhave\t%v", len(expect), len(result))
		return
	}
	for i := 0; i < len(expect); i++ {
		if !reflect.DeepEqual(expect[i], result[i]) {
			t.Errorf("[%d] results not match:\nwant:\t%#v\nhave\t%#v", i, expect[i], result[i])
		}
	}
}

func TestGetByUser_ErrNoPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cln := NewMockMongoCollection(ctrl)
	cs := NewMockMongoCursor(ctrl)

	service := &PostRepositoryMongo{coll: cln}

	expect := fmt.Errorf("some err")

	cln.EXPECT().
		Find(emptyCtx, bson.M{"author.username": usr1.ID}).
		Return(cs, nil)
	cs.EXPECT().
		All(emptyCtx, &[]*Post{}).
		Return(expect)

	result, err := service.GetByUser(usr1.ID)

	if result != nil {
		t.Errorf("unexpected result: %#v", result)
	}
	if !errors.Is(err, expect) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, err)
	}
}
