package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"asperitas/internal/session"
	"asperitas/internal/user"

	"go.uber.org/zap"
)

type UserHandler struct {
	Sess   session.SessionManager
	Repo   user.UserRepo
	Logger *zap.SugaredLogger
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds user.Credentials
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		WriteAndLogErr(w, err, h.Logger, "decode json err")
		return
	}
	usr, err := h.Repo.Authorize(creds.Username, creds.Password)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "authorize err")
		return
	}
	sess, err := h.Sess.Create(usr)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "create session err")
		return
	}
	logStr := fmt.Sprintf("logged user: username=%s id=%s", usr.Username, usr.ID)
	WriteAndLogData(w, sess, h.Logger, logStr)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds user.Credentials
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		WriteAndLogErr(w, err, h.Logger, "decode json err")
		return
	}
	usr, err := h.Repo.SignUp(creds.Username, creds.Password)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "sign up err")
		return
	}
	sess, err := h.Sess.Create(usr)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "create session err")
		return
	}
	logStr := fmt.Sprintf("registered user: username=%s id=%s", usr.Username, usr.ID)
	WriteAndLogData(w, sess, h.Logger, logStr)
}
