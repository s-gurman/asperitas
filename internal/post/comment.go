package post

import (
	"fmt"
	"time"

	"asperitas/internal/errs"
	"asperitas/internal/user"
	"asperitas/pkg/rand"
)

type Comment struct {
	Created       time.Time `json:"-"`
	CreatedFormat string    `json:"created"`
	Author        user.User `json:"author"`
	Body          string    `json:"body"`
	ID            string    `json:"id"`
}

type CommentList []*Comment

func NewComment(usr user.User, body string) *Comment {
	t := time.Now()
	return &Comment{
		Created:       t,
		CreatedFormat: t.Format(time.RFC3339Nano),
		Author:        usr,
		Body:          body,
		ID:            rand.GetRandID(),
	}
}

func (list *CommentList) Add(comm *Comment) error {
	if list == nil || *list == nil {
		return fmt.Errorf("nil comment list")
	}
	*list = append(*list, comm)
	return nil
}

func (list *CommentList) Delete(id, reqID string) error {
	if list == nil || *list == nil {
		return fmt.Errorf("nil comment list")
	}
	i := -1
	for idx, comm := range *list {
		if comm.ID == id {
			i = idx
			break
		}
	}
	if i < 0 {
		return errs.MsgError{Msg: "comment not found", Status: 404}
	}
	if (*list)[i].Author.ID != reqID {
		return errs.MsgError{Msg: "unauthorized", Status: 401}
	}
	(*list)[i] = (*list)[len(*list)-1]
	*list = (*list)[:len(*list)-1]
	return nil
}
