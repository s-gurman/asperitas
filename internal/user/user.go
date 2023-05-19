package user

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Username string `json:"username"`
	ID       string `json:"id"`
	Password string `json:"-"`
}

type UserRepo interface {
	Authorize(username, passw string) (*User, error)
	SignUp(username, passw string) (*User, error)
}
