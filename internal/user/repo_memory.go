package user

import (
	"sync"

	"asperitas/internal/errs"
	"asperitas/pkg/rand"
)

type UserMemoryRepository struct {
	data map[string]*User
	mu   *sync.RWMutex
}

func NewMemoryRepo() *UserMemoryRepository {
	return &UserMemoryRepository{
		data: map[string]*User{
			"admin1": {
				ID:       "id_admin1",
				Username: "admin1",
				Password: "adminadmin",
			},
			"admin2": {
				ID:       "id_admin2",
				Username: "admin2",
				Password: "adminadmin",
			},
		},
		mu: &sync.RWMutex{},
	}
}

func (repo *UserMemoryRepository) Authorize(username, passw string) (*User, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	usr, ok := repo.data[username]
	if !ok {
		return nil, errs.MsgError{Msg: "user not found", Status: 401}
	}
	if usr.Password != passw {
		return nil, errs.MsgError{Msg: "invalid password", Status: 401}
	}
	return usr, nil
}

func (repo *UserMemoryRepository) SignUp(username, passw string) (*User, error) {
	repo.mu.RLock()
	_, exist := repo.data[username]
	repo.mu.RUnlock()
	if exist {
		return nil, errs.DetailErrors{Errors: []errs.DetailError{
			{
				Location: "body",
				Param:    "username",
				Value:    username,
				Msg:      "already exists",
			},
		}, Status: 422}
	}
	usr := &User{
		ID:       rand.GetRandID(),
		Username: username,
		Password: passw,
	}
	repo.mu.Lock()
	repo.data[username] = usr
	repo.mu.Unlock()
	return usr, nil
}
