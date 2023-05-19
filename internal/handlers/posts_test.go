package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"asperitas/internal/errs"
	"asperitas/internal/post"
	"asperitas/internal/session"
	"asperitas/internal/user"
	"asperitas/pkg/rand"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"
)

var (
	usr1   = user.User{Username: "admin1", ID: "id_admin1", Password: "passw"}
	usr2   = user.User{Username: "admin2", ID: "id_admin2", Password: "passw"}
	randID = rand.GetRandID()
)

func getMockPostService(t *testing.T) (*PostHandler, *session.MockSessionManager, *post.MockPostRepo) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mng := session.NewMockSessionManager(ctrl)
	db := post.NewMockPostRepo(ctrl)
	return &PostHandler{
		Sess:   mng,
		Repo:   db,
		Logger: zap.NewNop().Sugar(),
	}, mng, db
}

func TestListPosts_OK(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := []*post.Post{post.NewPost(usr1), post.NewPost(usr2)}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/posts/", nil)
	w := httptest.NewRecorder()

	db.EXPECT().
		GetAll().
		Return(expect, nil)

	service.ListPosts(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%#v\nhave\t%#v", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestListPosts_GetErr(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/posts/", nil)
	w := httptest.NewRecorder()

	db.EXPECT().
		GetAll().
		Return(nil, fmt.Errorf("some err"))

	service.ListPosts(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%#v\nhave\t%#v", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestCreatePost_OK(t *testing.T) {
	service, mng, db := getMockPostService(t)

	reqBytes := []byte(`{"category":"music","type":"text","title":"some title","text":"some text"}`)
	expect := post.NewPost(usr1)
	json.Unmarshal(reqBytes, expect)      // nolint:errcheck
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBuffer(reqBytes)
	req := httptest.NewRequest("POST", "/api/posts", reqBody)
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		AddPost(gomock.Any()).
		Return(nil)

	service.CreatePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%#v\nhave\t%#v", 200, resp.StatusCode)
	}
	patch, err := jsondiff.CompareJSON(body, expectBody)
	if err != nil {
		t.Errorf("bad resp body: %s\n", body)
	}
	for _, op := range patch {
		if op.Path.String() != "/created" && op.Path.String() != "/id" {
			t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
			return
		}
	}
}

func TestCreatePost_AuthErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("POST", "/api/posts", reqBody)
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(user.User{}, expect)

	service.CreatePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%#v\nhave\t%#v", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestCreatePost_DecodeErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`bad json`)
	req := httptest.NewRequest("POST", "/api/posts", reqBody)
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)

	service.CreatePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestCreatePost_AddErr(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("POST", "/api/posts", reqBody)
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		AddPost(gomock.Any()).
		Return(fmt.Errorf("some err"))

	service.CreatePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestListPostsByCategory_OK(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := []*post.Post{post.NewPost(usr1), post.NewPost(usr2)}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/posts/{categoryName}", nil)
	req = mux.SetURLVars(req, map[string]string{"categoryName": "music"})
	w := httptest.NewRecorder()

	db.EXPECT().
		GetByCategory("music").
		Return(expect, nil)

	service.ListPostsByCategory(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestListPostsByCategory_GetErr(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/posts/{categoryName}", nil)
	req = mux.SetURLVars(req, map[string]string{"categoryName": "music"})
	w := httptest.NewRecorder()

	db.EXPECT().
		GetByCategory("music").
		Return(nil, fmt.Errorf("some err"))

	service.ListPostsByCategory(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestShowPost_OK(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := post.NewPost(usr1)
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": expect.ID})
	w := httptest.NewRecorder()

	db.EXPECT().
		GetByID(expect.ID).
		Return(expect, nil)

	service.ShowPost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestShowPost_InvalidErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid post id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": "invalid id"})
	w := httptest.NewRecorder()

	service.ShowPost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestShowPost_GetErr(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	db.EXPECT().
		GetByID(randID).
		Return(nil, expect)

	service.ShowPost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDeletePost_OK(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "success", Status: 200}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		DeletePost(randID, usr1.ID).
		Return(expect)

	service.DeletePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDeletePost_InvalidErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid post id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": "invalid id"})
	w := httptest.NewRecorder()

	service.DeletePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDeletePost_AuthErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(user.User{}, expect)

	service.DeletePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestDeletePost_DeleteErr(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		DeletePost(randID, usr1.ID).
		Return(expect)

	service.DeletePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestCreateComment_OK(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := post.NewPost(usr1)
	comm := post.NewComment(usr2, "some comment")
	expect.Comments.Add(comm)             // nolint:errcheck
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`{"comment":"some comment"}`)
	req := httptest.NewRequest("POST", "/api/post/{postID}", reqBody)
	req = mux.SetURLVars(req, map[string]string{"postID": expect.ID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr2, nil)
	db.EXPECT().
		AddComment(expect.ID, gomock.Any()).
		Return(expect, nil)

	service.CreateComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestCreateComment_InvalidErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid post id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("POST", "/api/post/{postID}", reqBody)
	req = mux.SetURLVars(req, map[string]string{"postID": "invalid id"})
	w := httptest.NewRecorder()

	service.CreateComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestCreateComment_AuthErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("POST", "/api/post/{postID}", reqBody)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(user.User{}, expect)

	service.CreateComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestCreateComment_DecodeErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`bad json`)
	req := httptest.NewRequest("POST", "/api/post/{postID}", reqBody)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)

	service.CreateComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestCreateComment_EmptyCommentErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.DetailErrors{Errors: []errs.DetailError{
		{
			Location: "body",
			Param:    "comment",
			Msg:      "is required",
		},
	}, Status: 422}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`{"comment":""}`)
	req := httptest.NewRequest("POST", "/api/post/{postID}", reqBody)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)

	service.CreateComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestCreateComment_AddErr(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`{"comment":"some comment"}`)
	req := httptest.NewRequest("POST", "/api/post/{postID}", reqBody)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		AddComment(randID, gomock.Any()).
		Return(nil, expect)

	service.CreateComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestDeleteComment_OK(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := post.NewPost(usr1)
	comm := post.NewComment(usr2, "some comment")
	expect.Comments.Add(comm)                // nolint:errcheck
	expect.Comments.Delete(comm.ID, usr2.ID) // nolint:errcheck
	expectBody, _ := json.Marshal(expect)    // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}/{commentID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": expect.ID, "commentID": comm.ID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr2, nil)
	db.EXPECT().
		DeleteComment(expect.ID, comm.ID, usr2.ID).
		Return(expect, nil)

	service.DeleteComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDeleteComment_InvalidPostErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid post id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}/{commentID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": "invalid id", "commentID": randID})
	w := httptest.NewRecorder()

	service.DeleteComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDeleteComment_InvalidCommentErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid comment id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}/{commentID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID, "commentID": "invalid id"})
	w := httptest.NewRecorder()

	service.DeleteComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDeleteComment_AuthErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}/{commentID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID, "commentID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(user.User{}, expect)

	service.DeleteComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestDeleteComment_DeleteErr(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("DELETE", "/api/post/{postID}/{commentID}", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID, "commentID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		DeleteComment(randID, randID, usr1.ID).
		Return(nil, expect)

	service.DeleteComment(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestUpvotePost_OK(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := post.NewPost(usr1)
	expect.Votes.Upvote(usr2.ID)          // nolint:errcheck
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/upvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": expect.ID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr2, nil)
	db.EXPECT().
		UpvotePost(expect.ID, usr2.ID).
		Return(expect, nil)

	service.UpvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestUpvotePost_InvalidErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid post id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/upvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": "invalid id"})
	w := httptest.NewRecorder()

	service.UpvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestUpvotePost_AuthErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/upvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(user.User{}, expect)

	service.UpvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestUpvotePost_UpvoteErr(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/upvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		UpvotePost(randID, usr1.ID).
		Return(nil, expect)

	service.UpvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDownvotePost_OK(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := post.NewPost(usr1)
	expect.Votes.Downvote(usr2.ID)        // nolint:errcheck
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/downvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": expect.ID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr2, nil)
	db.EXPECT().
		DownvotePost(expect.ID, usr2.ID).
		Return(expect, nil)

	service.DownvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDownvotePost_InvalidErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid post id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/downvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": "invalid id"})
	w := httptest.NewRecorder()

	service.DownvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestDownvotePost_AuthErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/downvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(user.User{}, expect)

	service.DownvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestDownvotePost_DownvoteErr(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/downvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		DownvotePost(randID, usr1.ID).
		Return(nil, expect)

	service.DownvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestUnvotePost_OK(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := post.NewPost(usr1)
	expect.Votes.Unvote(usr1.ID)          // nolint:errcheck
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/unvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": expect.ID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		UnvotePost(expect.ID, usr1.ID).
		Return(expect, nil)

	service.UnvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestUnvotePost_InvalidErr(t *testing.T) {
	service, _, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "invalid post id", Status: 400}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/unvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": "invalid id"})
	w := httptest.NewRecorder()

	service.UnvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestUnvotePost_AuthErr(t *testing.T) {
	service, mng, _ := getMockPostService(t)

	expect := errs.MsgError{Msg: "unauthorized", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/unvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(user.User{}, expect)

	service.UnvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestUnvotePost_UnvoteErr(t *testing.T) {
	service, mng, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "post not found", Status: 404}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/post/{postID}/unvote", nil)
	req = mux.SetURLVars(req, map[string]string{"postID": randID})
	w := httptest.NewRecorder()

	mng.EXPECT().
		Check(gomock.Any()).
		Return(usr1, nil)
	db.EXPECT().
		UnvotePost(randID, usr1.ID).
		Return(nil, expect)

	service.UnvotePost(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestListPostsByUser_OK(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := []*post.Post{post.NewPost(usr1), post.NewPost(usr2)}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/user/{userName}", nil)
	req = mux.SetURLVars(req, map[string]string{"userName": "grant"})
	w := httptest.NewRecorder()

	db.EXPECT().
		GetByUser("grant").
		Return(expect, nil)

	service.ListPostsByUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}

func TestListPostsByUser_GetErr(t *testing.T) {
	service, _, db := getMockPostService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	req := httptest.NewRequest("GET", "/api/user/{userName}", nil)
	req = mux.SetURLVars(req, map[string]string{"userName": "grant"})
	w := httptest.NewRecorder()

	db.EXPECT().
		GetByUser("grant").
		Return(nil, fmt.Errorf("some err"))

	service.ListPostsByUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != expect.Status {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", expect.Status, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\n\nwant:\t%s\n\nhave\t%s", expectBody, body)
	}
}
