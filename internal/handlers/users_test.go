package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"asperitas/internal/errs"
	"asperitas/internal/session"
	"asperitas/internal/user"
	"asperitas/pkg/rand"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

var (
	usr   = &user.User{ID: rand.GetRandID(), Username: "admin", Password: "passw"}
	creds = user.Credentials{Username: "admin", Password: "passw"}
)

func getMockUserService(t *testing.T) (*UserHandler, *session.MockSessionManager, *user.MockUserRepo) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mng := session.NewMockSessionManager(ctrl)
	db := user.NewMockUserRepo(ctrl)
	return &UserHandler{
		Sess:   mng,
		Repo:   db,
		Logger: zap.NewNop().Sugar(),
	}, mng, db
}

func getCredsBuffer() *bytes.Buffer {
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(creds) // nolint:errcheck
	return buff
}

func TestLogin_OK(t *testing.T) {
	service, mng, db := getMockUserService(t)

	expect := session.Session{Token: "some token"}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/login", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		Authorize(creds.Username, creds.Password).
		Return(usr, nil)
	mng.EXPECT().
		Create(usr).
		Return(expect, nil)

	service.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestLogin_DecodeErr(t *testing.T) {
	service, _, _ := getMockUserService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`bad json`)
	req := httptest.NewRequest("POST", "/api/login", reqBody)
	w := httptest.NewRecorder()

	service.Login(w, req)

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

func TestLogin_AuthErr(t *testing.T) {
	service, _, db := getMockUserService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/login", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		Authorize(creds.Username, creds.Password).
		Return(nil, fmt.Errorf("mysql scan err"))

	service.Login(w, req)

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

func TestLogin_AuthMsgErr(t *testing.T) {
	service, _, db := getMockUserService(t)

	expect := errs.MsgError{Msg: "user not found", Status: 401}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/login", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		Authorize(creds.Username, creds.Password).
		Return(nil, expect)

	service.Login(w, req)

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

func TestLogin_SessCreateErr(t *testing.T) {
	service, mng, db := getMockUserService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/login", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		Authorize(creds.Username, creds.Password).
		Return(usr, nil)
	mng.EXPECT().
		Create(usr).
		Return(session.Session{}, fmt.Errorf("mysql exec err"))

	service.Login(w, req)

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

func TestRegister_OK(t *testing.T) {
	service, mng, db := getMockUserService(t)

	expect := session.Session{Token: "some token"}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/register", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		SignUp(creds.Username, creds.Password).
		Return(usr, nil)
	mng.EXPECT().
		Create(usr).
		Return(expect, nil)

	service.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body) // nolint:errcheck

	if resp.StatusCode != 200 {
		t.Errorf("wrong status code:\nwant:\t%d\nhave\t%d", 200, resp.StatusCode)
	}
	if string(body) != string(expectBody) {
		t.Errorf("results not match:\nwant:\t%s\nhave\t%s", expectBody, body)
	}
}

func TestRegister_DecodeErr(t *testing.T) {
	service, _, _ := getMockUserService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := bytes.NewBufferString(`bad json`)
	req := httptest.NewRequest("POST", "/api/register", reqBody)
	w := httptest.NewRecorder()

	service.Register(w, req)

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

func TestRegister_SignUpErr(t *testing.T) {
	service, _, db := getMockUserService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/register", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		SignUp(creds.Username, creds.Password).
		Return(nil, fmt.Errorf("mysql scan err"))

	service.Register(w, req)

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

func TestRegister_SignUpMsgErr(t *testing.T) {
	service, _, db := getMockUserService(t)

	expect := errs.DetailErrors{
		Errors: []errs.DetailError{
			{
				Location: "body",
				Param:    "username",
				Value:    "admin",
				Msg:      "already exists",
			},
		},
		Status: 422,
	}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/register", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		SignUp(creds.Username, creds.Password).
		Return(nil, expect)

	service.Register(w, req)

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

func TestRegister_SessCreateErr(t *testing.T) {
	service, mng, db := getMockUserService(t)

	expect := errs.MsgError{Msg: "internal server error", Status: 500}
	expectBody, _ := json.Marshal(expect) // nolint:errcheck

	reqBody := getCredsBuffer()
	req := httptest.NewRequest("POST", "/api/register", reqBody)
	w := httptest.NewRecorder()

	db.EXPECT().
		SignUp(creds.Username, creds.Password).
		Return(usr, nil)
	mng.EXPECT().
		Create(usr).
		Return(session.Session{}, fmt.Errorf("mysql exec err"))

	service.Register(w, req)

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
