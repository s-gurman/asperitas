package post

import (
	"time"

	"asperitas/internal/user"
	"asperitas/pkg/rand"
)

type PostType string

const (
	Text PostType = "text"
	Link PostType = "link"
)

type PostCategory string

const (
	Music       PostCategory = "music"
	Funny       PostCategory = "funny"
	Videos      PostCategory = "videos"
	Programming PostCategory = "programming"
	News        PostCategory = "news"
	Fashion     PostCategory = "fashion"
)

type Post struct {
	Score         int64        `json:"score"`
	Views         uint32       `json:"views"`
	Type          PostType     `json:"type"`
	Title         string       `json:"title"`
	URL           string       `json:"url,omitempty"`
	Author        user.User    `json:"author"`
	Category      PostCategory `json:"category"`
	Text          string       `json:"text,omitempty"`
	Votes         VoteList     `json:"votes"`
	Comments      CommentList  `json:"comments"`
	Created       time.Time    `json:"-"`
	CreatedFormat string       `json:"created"`
	LikesPercent  int          `json:"upvotePercentage"`
	ID            string       `json:"id"`
}

type PostRepo interface {
	GetAll() ([]*Post, error)
	AddPost(post *Post) error
	GetByCategory(category string) ([]*Post, error)
	GetByID(postID string) (*Post, error)
	DeletePost(postID, userID string) error
	AddComment(postID string, comment *Comment) (*Post, error)
	DeleteComment(postID, commentID, userID string) (*Post, error)
	UpvotePost(postID, userID string) (*Post, error)
	DownvotePost(postID, userID string) (*Post, error)
	UnvotePost(postID, userID string) (*Post, error)
	GetByUser(username string) ([]*Post, error)
}

func (p *Post) updatePostScore() {
	p.Score = int64(2*p.Votes.LikesCount - len(p.Votes.List))
	if len(p.Votes.List) == 0 {
		p.LikesPercent = 0
	} else {
		p.LikesPercent = p.Votes.LikesCount * 100 / len(p.Votes.List)
	}
}

func NewPost(usr user.User) *Post {
	t := time.Now()
	return &Post{
		Score:         1,
		Views:         0,
		Author:        usr,
		Created:       t,
		CreatedFormat: t.Format(time.RFC3339Nano),
		Votes:         NewVoteList(usr.ID),
		Comments:      make(CommentList, 0, 1000),
		LikesPercent:  100,
		ID:            rand.GetRandID(),
	}
}
