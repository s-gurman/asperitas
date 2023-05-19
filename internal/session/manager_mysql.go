package session

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"asperitas/internal/errs"
	"asperitas/internal/user"
)

type SessionManagerMySQL struct {
	db *sql.DB
}

func NewManagerMySQL(addr string) (*SessionManagerMySQL, error) {
	mySQL, err := sql.Open("mysql", addr)
	if err != nil {
		return nil, fmt.Errorf("mysql open err: %w", err)
	}
	err = mySQL.Ping()
	if err != nil {
		return nil, fmt.Errorf("mysql connect err: %w", err)
	}
	return &SessionManagerMySQL{db: mySQL}, nil
}

func (sm *SessionManagerMySQL) Create(usr *user.User) (Session, error) {
	sess, err := NewSession(usr)
	if err != nil {
		return Session{}, err
	}
	_, err = sm.db.Exec(
		"INSERT INTO `sessions` (`session_id`) VALUES (?)",
		sess.ID,
	)
	if err != nil {
		return Session{}, fmt.Errorf("mysql exec insert err: %w", err)
	}
	return sess, nil
}

func (sm *SessionManagerMySQL) Check(authHeader string) (user.User, error) {
	authFields := strings.Fields(authHeader) // "Authorization": "Bearer <token>"
	if len(authFields) != 2 || authFields[0] != "Bearer" {
		return user.User{}, errs.MsgError{Msg: "unauthorized", Status: 401}
	}
	claims, err := ExtractJwtClaims(authFields[1])
	if err != nil {
		return user.User{}, err
	}
	err = sm.db.
		QueryRow("SELECT `session_id` FROM `sessions` WHERE `session_id` = ?", claims.SessionID).
		Scan(&claims.SessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, errs.MsgError{Msg: "unauthorized", Status: 401}
	} else if err != nil {
		return user.User{}, fmt.Errorf("mysql scan err: %w", err)
	}
	return claims.User, nil
}
