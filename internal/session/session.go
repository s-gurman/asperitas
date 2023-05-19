package session

import (
	"fmt"
	"time"

	"asperitas/internal/errs"
	"asperitas/internal/user"
	"asperitas/pkg/rand"

	jwt "github.com/golang-jwt/jwt/v4"
)

const jwtKey = "secretKey"

type Claims struct {
	User      user.User `json:"user"`
	SessionID string    `json:"session_id"`
	jwt.RegisteredClaims
}

type Session struct {
	Token string `json:"token"`
	ID    string `json:"-"`
}

type SessionManager interface {
	Create(usr *user.User) (Session, error)
	Check(authHeader string) (user.User, error)
}

func NewSession(usr *user.User) (Session, error) {
	if usr == nil {
		return Session{}, fmt.Errorf("nil input user")
	}
	randID := rand.GetRandID()
	claims := &Claims{
		User:      *usr,
		SessionID: randID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return Session{}, err
	}
	return Session{Token: tokenString, ID: randID}, nil
}

func ExtractJwtClaims(tokenString string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(*jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		return claims, fmt.Errorf("jwt parse err: %w", err)
	}
	if !token.Valid {
		return claims, errs.MsgError{Msg: "unauthorized", Status: 401}
	}
	return claims, nil
}
