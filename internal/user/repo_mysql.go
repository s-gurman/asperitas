package user

import (
	"database/sql"
	"errors"
	"fmt"

	"asperitas/internal/errs"
	"asperitas/pkg/rand"

	_ "github.com/go-sql-driver/mysql"
)

type UserRepositoryMySQL struct {
	db *sql.DB
}

func NewRepoMySQL(addr string) (*UserRepositoryMySQL, error) {
	mySQL, err := sql.Open("mysql", addr)
	if err != nil {
		return nil, fmt.Errorf("mysql open err: %w", err)
	}
	err = mySQL.Ping()
	if err != nil {
		return nil, fmt.Errorf("mysql connect err: %w", err)
	}
	return &UserRepositoryMySQL{db: mySQL}, nil
}

func (repo *UserRepositoryMySQL) Authorize(username, passw string) (*User, error) {
	usr := &User{Username: username}
	err := repo.db.
		QueryRow("SELECT `id`, `password` FROM `users` WHERE `username` = ?", username).
		Scan(&usr.ID, &usr.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errs.MsgError{Msg: "user not found", Status: 401}
	} else if err != nil {
		return nil, fmt.Errorf("mysql scan err: %w", err)
	}
	if usr.Password != passw {
		return nil, errs.MsgError{Msg: "invalid password", Status: 401}
	}
	return usr, nil
}

func (repo *UserRepositoryMySQL) SignUp(username, passw string) (*User, error) {
	usr := &User{Username: username}
	err := repo.db.
		QueryRow("SELECT `id`, `password` FROM `users` WHERE `username` = ?", username).
		Scan(&usr.ID, &usr.Password)
	if errors.Is(err, sql.ErrNoRows) {
		usr.ID = rand.GetRandID()
		usr.Password = passw
		if _, err = repo.db.Exec(
			"INSERT INTO `users` (`username`, `id`, `password`) VALUES (?, ?, ?)",
			username,
			usr.ID,
			passw,
		); err != nil {
			return nil, fmt.Errorf("mysql exec insert err: %w", err)
		}
		return usr, nil
	} else if err != nil {
		return nil, fmt.Errorf("mysql scan err: %w", err)
	}
	return nil, errs.DetailErrors{Errors: []errs.DetailError{
		{
			Location: "body",
			Param:    "username",
			Value:    username,
			Msg:      "already exists",
		},
	}, Status: 422}
}
