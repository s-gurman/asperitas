package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"asperitas/internal/errs"
	"asperitas/internal/post"
	"asperitas/internal/session"
	"asperitas/internal/user"
	"asperitas/pkg/rand"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PostHandler struct {
	Sess   session.SessionManager
	Repo   post.PostRepo
	Logger *zap.SugaredLogger
}

// Проверяет валидность входящего ID по длине до похода в репу
func isValid(reqID string, respMsg string, w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) (string, bool) {
	id := mux.Vars(r)[reqID]
	if len(id) != rand.LengthOfID {
		err := errs.MsgError{Msg: respMsg, Status: http.StatusBadRequest}
		WriteAndLogErr(w, err, logger, fmt.Sprintf("%s valid err", reqID))
		return "", false
	}
	return id, true
}

// Проверяет валидность сессии по полученному jwt-токену и сохраненному в базе session_id
func sessionCheck(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, sm session.SessionManager) (user.User, bool) {
	usr, err := sm.Check(r.Header.Get("Authorization"))
	if err != nil {
		WriteAndLogErr(w, err, logger, "session check err")
		return user.User{}, false
	}
	return usr, true
}

func (h *PostHandler) ListPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.Repo.GetAll()
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "get all posts err")
		return
	}
	WriteAndLogData(w, posts, h.Logger, "listed all posts")
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	usr, ok := sessionCheck(w, r, h.Logger, h.Sess)
	if !ok {
		return
	}
	p := post.NewPost(usr)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(p); err != nil {
		WriteAndLogErr(w, err, h.Logger, "decode json err")
		return
	}
	if err := h.Repo.AddPost(p); err != nil {
		WriteAndLogErr(w, err, h.Logger, "add post err")
		return
	}
	logStr := fmt.Sprintf("created post: id=%s", p.ID)
	WriteAndLogData(w, p, h.Logger, logStr)
}

func (h *PostHandler) ListPostsByCategory(w http.ResponseWriter, r *http.Request) {
	category := mux.Vars(r)["categoryName"]
	posts, err := h.Repo.GetByCategory(category)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "get posts by category err")
		return
	}
	logStr := fmt.Sprintf("listed posts by: category=%s", category)
	WriteAndLogData(w, posts, h.Logger, logStr)
}

func (h *PostHandler) ShowPost(w http.ResponseWriter, r *http.Request) {
	id, ok := isValid("postID", "invalid post id", w, r, h.Logger)
	if !ok {
		return
	}
	p, err := h.Repo.GetByID(id)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "get post by id err")
		return
	}
	logStr := fmt.Sprintf("showed post: id=%s", id)
	WriteAndLogData(w, p, h.Logger, logStr)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, ok := isValid("postID", "invalid post id", w, r, h.Logger)
	if !ok {
		return
	}
	usr, ok := sessionCheck(w, r, h.Logger, h.Sess)
	if !ok {
		return
	}
	err := h.Repo.DeletePost(postID, usr.ID)
	msgErr, ok := err.(errs.MsgError)
	if ok && msgErr.Status == http.StatusOK {
		logStr := fmt.Sprintf("deleted post: id=%s", postID)
		WriteAndLogData(w, msgErr, h.Logger, logStr)
		return
	}
	WriteAndLogErr(w, err, h.Logger, "delete post err")
}

func (h *PostHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	postID, ok := isValid("postID", "invalid post id", w, r, h.Logger)
	if !ok {
		return
	}
	usr, ok := sessionCheck(w, r, h.Logger, h.Sess)
	if !ok {
		return
	}
	defer r.Body.Close()
	reqBody := struct{ Comment string }{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		WriteAndLogErr(w, err, h.Logger, "decode json err")
		return
	}
	if reqBody.Comment == "" {
		err := errs.DetailErrors{Errors: []errs.DetailError{
			{
				Location: "body",
				Param:    "comment",
				Msg:      "is required",
			},
		}, Status: 422}
		WriteAndLogErr(w, err, h.Logger, "create empty comment err")
		return
	}
	comm := post.NewComment(usr, reqBody.Comment)
	p, err := h.Repo.AddComment(postID, comm)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "get post by id err")
		return
	}
	logStr := fmt.Sprintf("created comment: id=%s", comm.ID)
	WriteAndLogData(w, p, h.Logger, logStr)
}

func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	postID, ok := isValid("postID", "invalid post id", w, r, h.Logger)
	if !ok {
		return
	}
	commID, ok := isValid("commentID", "invalid comment id", w, r, h.Logger)
	if !ok {
		return
	}
	usr, ok := sessionCheck(w, r, h.Logger, h.Sess)
	if !ok {
		return
	}
	p, err := h.Repo.DeleteComment(postID, commID, usr.ID)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "delete comment err")
		return
	}
	logStr := fmt.Sprintf("deleted comment: id=%s", commID)
	WriteAndLogData(w, p, h.Logger, logStr)
}

func (h *PostHandler) UpvotePost(w http.ResponseWriter, r *http.Request) {
	postID, ok := isValid("postID", "invalid post id", w, r, h.Logger)
	if !ok {
		return
	}
	usr, ok := sessionCheck(w, r, h.Logger, h.Sess)
	if !ok {
		return
	}
	p, err := h.Repo.UpvotePost(postID, usr.ID)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "upvote post err")
		return
	}
	logStr := fmt.Sprintf("upvoted post: id=%s", postID)
	WriteAndLogData(w, p, h.Logger, logStr)
}

func (h *PostHandler) DownvotePost(w http.ResponseWriter, r *http.Request) {
	postID, ok := isValid("postID", "invalid post id", w, r, h.Logger)
	if !ok {
		return
	}
	usr, ok := sessionCheck(w, r, h.Logger, h.Sess)
	if !ok {
		return
	}
	p, err := h.Repo.DownvotePost(postID, usr.ID)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "downvote post err")
		return
	}
	logStr := fmt.Sprintf("downvoted post: id=%s", postID)
	WriteAndLogData(w, p, h.Logger, logStr)
}

func (h *PostHandler) UnvotePost(w http.ResponseWriter, r *http.Request) {
	postID, ok := isValid("postID", "invalid post id", w, r, h.Logger)
	if !ok {
		return
	}
	usr, ok := sessionCheck(w, r, h.Logger, h.Sess)
	if !ok {
		return
	}
	p, err := h.Repo.UnvotePost(postID, usr.ID)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "unvote post err")
		return
	}
	logStr := fmt.Sprintf("unvoted post: id=%s", postID)
	WriteAndLogData(w, p, h.Logger, logStr)
}

func (h *PostHandler) ListPostsByUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["userName"]
	posts, err := h.Repo.GetByUser(username)
	if err != nil {
		WriteAndLogErr(w, err, h.Logger, "get posts by user err")
		return
	}
	logStr := fmt.Sprintf("listed posts by: username=%s", username)
	WriteAndLogData(w, posts, h.Logger, logStr)
}
