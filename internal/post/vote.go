package post

import (
	"encoding/json"
	"fmt"
)

type VoteValue int

const (
	Like    VoteValue = 1
	Dislike VoteValue = -1
)

type Vote struct {
	UserID string    `json:"user"`
	Value  VoteValue `json:"vote"`
}

type VoteList struct {
	List       []Vote
	LikesCount int
}

func (l *VoteList) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.List)
}

func (l *VoteList) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &l.List)
}

func NewVoteList(id string) VoteList {
	votes := make([]Vote, 0, 1000)
	votes = append(votes, Vote{Value: Like, UserID: id})
	return VoteList{
		List:       votes,
		LikesCount: 1,
	}
}

func (l *VoteList) Upvote(userID string) error {
	if l == nil || l.List == nil {
		return fmt.Errorf("nil vote list")
	}
	l.LikesCount++
	for i, vote := range l.List {
		if vote.UserID == userID {
			l.List[i].Value = Like
			return nil
		}
	}
	l.List = append(l.List, Vote{UserID: userID, Value: Like})
	return nil
}

func (l *VoteList) Downvote(userID string) error {
	if l == nil || l.List == nil {
		return fmt.Errorf("nil vote list")
	}
	for i, vote := range l.List {
		if vote.UserID == userID {
			if vote.Value == Like {
				l.LikesCount--
			}
			l.List[i].Value = Dislike
			return nil
		}
	}
	l.List = append(l.List, Vote{UserID: userID, Value: Dislike})
	return nil
}

func (l *VoteList) Unvote(userID string) error {
	if l == nil || l.List == nil {
		return fmt.Errorf("nil vote list")
	}
	for i, vote := range l.List {
		if vote.UserID == userID {
			if vote.Value == Like {
				l.LikesCount--
			}
			l.List[i] = (l.List)[len(l.List)-1]
			l.List = (l.List)[:len(l.List)-1]
			return nil
		}
	}
	return nil
}
