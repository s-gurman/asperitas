package post

import (
	"sort"
	"sync"

	"asperitas/internal/errs"
)

type PostMemoryRepository struct {
	data []*Post
	mu   *sync.RWMutex
}

func NewMemoryRepo() *PostMemoryRepository {
	return &PostMemoryRepository{
		data: make([]*Post, 0, 1000),
		mu:   &sync.RWMutex{},
	}
}

func (repo *PostMemoryRepository) GetAll() ([]*Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	result := repo.data
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Created.Before(result[j].Created)
		}
		return result[i].Score > result[j].Score
	})
	return result, nil
}

func (repo *PostMemoryRepository) AddPost(p *Post) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.data = append(repo.data, p)
	return nil
}

func (repo *PostMemoryRepository) GetByCategory(categoryName string) ([]*Post, error) {
	category := PostCategory(categoryName)
	result := make([]*Post, 0, 1000)
	repo.mu.RLock()
	for _, post := range repo.data {
		if post.Category == category {
			result = append(result, post)
		}
	}
	repo.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Created.Before(result[j].Created)
		}
		return result[i].Score > result[j].Score
	})
	return result, nil
}

func (repo *PostMemoryRepository) GetByID(id string) (*Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, post := range repo.data {
		if post.ID == id {
			post.Views++
			return post, nil
		}
	}
	return nil, errs.MsgError{Msg: "post not found", Status: 404}
}

func (repo *PostMemoryRepository) DeletePost(postID, userID string) error {
	i := -1
	repo.mu.RLock()
	for idx, post := range repo.data {
		if post.ID == postID {
			i = idx
			break
		}
	}
	repo.mu.RUnlock()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if i < 0 {
		return errs.MsgError{Msg: "post not found", Status: 404}
	}
	if repo.data[i].Author.ID != userID {
		return errs.MsgError{Msg: "unauthorized", Status: 401}
	}
	repo.data[i] = repo.data[len(repo.data)-1]
	repo.data = repo.data[:len(repo.data)-1]
	return errs.MsgError{Msg: "success", Status: 200}
}

func (repo *PostMemoryRepository) AddComment(postID string, comm *Comment) (*Post, error) {
	var p *Post
	repo.mu.RLock()
	for _, post := range repo.data {
		if post.ID == postID {
			p = post
			break
		}
	}
	repo.mu.RUnlock()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if p == nil {
		return nil, errs.MsgError{Msg: "post not found", Status: 404}
	}
	if err := p.Comments.Add(comm); err != nil {
		return nil, err
	}
	return p, nil
}

func (repo *PostMemoryRepository) DeleteComment(postID, commID, userID string) (*Post, error) {
	var p *Post
	repo.mu.RLock()
	for _, post := range repo.data {
		if post.ID == postID {
			p = post
			break
		}
	}
	repo.mu.RUnlock()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if p == nil {
		return nil, errs.MsgError{Msg: "post not found", Status: 404}
	}
	err := p.Comments.Delete(commID, userID)
	return p, err
}

func (repo *PostMemoryRepository) UpvotePost(postID, userID string) (*Post, error) {
	var p *Post
	repo.mu.RLock()
	for _, post := range repo.data {
		if post.ID == postID {
			p = post
			break
		}
	}
	repo.mu.RUnlock()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if p == nil {
		return nil, errs.MsgError{Msg: "post not found", Status: 404}
	}
	err := p.Votes.Upvote(userID)
	if err == nil {
		p.Score = int64(2*p.Votes.LikesCount - len(p.Votes.List))
		if len(p.Votes.List) == 0 {
			p.LikesPercent = 0
		} else {
			p.LikesPercent = p.Votes.LikesCount * 100 / len(p.Votes.List)
		}
	}
	return p, err
}

func (repo *PostMemoryRepository) DownvotePost(postID, userID string) (*Post, error) {
	var p *Post
	repo.mu.RLock()
	for _, post := range repo.data {
		if post.ID == postID {
			p = post
			break
		}
	}
	repo.mu.RUnlock()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if p == nil {
		return nil, errs.MsgError{Msg: "post not found", Status: 404}
	}
	err := p.Votes.Downvote(userID)
	if err == nil {
		p.Score = int64(2*p.Votes.LikesCount - len(p.Votes.List))
		if len(p.Votes.List) == 0 {
			p.LikesPercent = 0
		} else {
			p.LikesPercent = p.Votes.LikesCount * 100 / len(p.Votes.List)
		}
	}
	return p, err
}

func (repo *PostMemoryRepository) UnvotePost(postID, userID string) (*Post, error) {
	var p *Post
	repo.mu.RLock()
	for _, post := range repo.data {
		if post.ID == postID {
			p = post
			break
		}
	}
	repo.mu.RUnlock()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if p == nil {
		return nil, errs.MsgError{Msg: "post not found", Status: 404}
	}
	err := p.Votes.Unvote(userID)
	if err == nil {
		p.Score = int64(2*p.Votes.LikesCount - len(p.Votes.List))
		if len(p.Votes.List) == 0 {
			p.LikesPercent = 0
		} else {
			p.LikesPercent = p.Votes.LikesCount * 100 / len(p.Votes.List)
		}
	}
	return p, err
}

func (repo *PostMemoryRepository) GetByUser(username string) ([]*Post, error) {
	result := make([]*Post, 0, 1000)
	repo.mu.RLock()
	for _, post := range repo.data {
		if post.Author.Username == username {
			result = append(result, post)
		}
	}
	repo.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool {
		return result[i].Created.After(result[j].Created)
	})
	return result, nil
}
