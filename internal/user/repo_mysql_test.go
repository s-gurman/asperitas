package user

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"asperitas/internal/errs"
	"asperitas/pkg/rand"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAuthorize_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := &User{Username: "admin", ID: rand.GetRandID(), Password: "passw"}
	rows := sqlmock.
		NewRows([]string{`id`, `password`}).
		AddRow(expect.ID, expect.Password)

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	usr, err := repo.Authorize(creds.Username, creds.Password)

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if !reflect.DeepEqual(expect, usr) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, usr)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthorize_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := errs.MsgError{Msg: "user not found", Status: 401}
	rows := sqlmock.NewRows([]string{`id`, `password`})

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	usr, err := repo.Authorize(creds.Username, creds.Password)

	if usr != nil {
		t.Errorf("unexpected usr: %#v", usr)
	}
	msgErr, ok := err.(errs.MsgError)
	if !ok {
		t.Errorf("unexpected err: \nwant:\t%#v\nhave\t%#v", expect, err)
	}
	if !reflect.DeepEqual(expect, msgErr) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, msgErr)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthorize_ScanErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := "mysql scan err"
	rows := sqlmock.
		NewRows([]string{`password`}).
		AddRow(creds.Password)

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	usr, err := repo.Authorize(creds.Username, creds.Password)

	if usr != nil {
		t.Errorf("unexpected usr: %#v", usr)
	}
	if err == nil || !strings.HasPrefix(err.Error(), expect) {
		t.Errorf("unexpected err:\nwant:\t%s\nhave\t%#v", expect, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthorize_WrongPassw(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := errs.MsgError{Msg: "invalid password", Status: 401}
	rows := sqlmock.
		NewRows([]string{`id`, `password`}).
		AddRow(creds.Username, "wrong passw")

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	usr, err := repo.Authorize(creds.Username, creds.Password)

	if usr != nil {
		t.Errorf("unexpected usr: %#v", usr)
	}
	msgErr, ok := err.(errs.MsgError)
	if !ok {
		t.Errorf("unexpected err: \nwant:\t%#v\nhave\t%#v", expect, err)
	}
	if !reflect.DeepEqual(expect, msgErr) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, msgErr)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSignUp_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := &User{Username: "admin", Password: "passw"}
	rows := sqlmock.NewRows([]string{`id`, `password`})

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	mock.
		ExpectExec("INSERT INTO `users`").
		WithArgs(creds.Username, sqlmock.AnyArg(), creds.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	usr, err := repo.SignUp(creds.Username, creds.Password)
	if usr != nil {
		expect.ID = usr.ID
	}

	if err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	if len(expect.ID) != rand.LengthOfID || !reflect.DeepEqual(expect, usr) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, usr)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSignUp_ExecErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := "mysql exec insert err"
	rows := sqlmock.NewRows([]string{`id`, `password`})

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	mock.
		ExpectExec("INSERT INTO `users`").
		WithArgs(creds.Username, sqlmock.AnyArg(), creds.Password).
		WillReturnError(fmt.Errorf("bad exec"))

	usr, err := repo.SignUp(creds.Username, creds.Password)

	if usr != nil {
		t.Errorf("unexpected usr: %#v", usr)
	}
	if err == nil || !strings.HasPrefix(err.Error(), expect) {
		t.Errorf("unexpected err:\nwant:\t%s\nhave\t%#v", expect, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSignUp_ScanErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := "mysql scan err"
	rows := sqlmock.
		NewRows([]string{`password`}).
		AddRow(creds.Password)

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	usr, err := repo.SignUp(creds.Username, creds.Password)

	if usr != nil {
		t.Errorf("unexpected usr: %#v", usr)
	}
	if err == nil || !strings.HasPrefix(err.Error(), expect) {
		t.Errorf("unexpected err:\nwant:\t%s\nhave\t%#v", expect, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSignUp_AlreadyExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock create err: %s", err)
	}
	defer db.Close()

	creds := Credentials{Username: "admin", Password: "passw"}
	repo := &UserRepositoryMySQL{db: db}

	expect := errs.DetailErrors{
		Errors: []errs.DetailError{
			{
				Location: "body",
				Param:    "username",
				Value:    creds.Username,
				Msg:      "already exists",
			},
		},
		Status: 422,
	}
	rows := sqlmock.
		NewRows([]string{`id`, `password`}).
		AddRow(rand.GetRandID(), "other passw")

	mock.
		ExpectQuery("SELECT `id`, `password` FROM `users` WHERE").
		WithArgs(creds.Username).
		WillReturnRows(rows)

	usr, err := repo.SignUp(creds.Username, creds.Password)

	if usr != nil {
		t.Errorf("unexpected usr: %#v", usr)
	}
	detailErrs, ok := err.(errs.DetailErrors)
	if !ok {
		t.Errorf("unexpected err: \nwant:\t%#v\nhave\t%#v", expect, err)
		return
	}
	if !reflect.DeepEqual(expect, detailErrs) {
		t.Errorf("results not match:\nwant:\t%#v\nhave\t%#v", expect, detailErrs)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
