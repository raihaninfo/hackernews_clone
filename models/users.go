package models

import (
	"errors"
	"time"

	"github.com/upper/db/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `db:"id,omitempty"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Password  string    `db:"password_hash"`
	CreatedAt time.Time `db:"created_at"`
	Activated bool      `db:"activated"`
}

type UserModel struct {
	db db.Session
}

func (u UserModel) Table() string {
	return "users"
}

func (m UserModel) Get(id int) (*User, error) {
	var u User

	err := m.db.Collection(m.Table()).Find(db.Cond{"id": id}).One(&u)
	if err != nil {
		if errors.Is(err, db.ErrNoMoreRows) {
			return nil, ErrNoMoreRows
		}
		return nil, err
	}

	return &u, nil
}

func (m UserModel) FindByEmail(email string) (*User, error) {
	var u User

	err := m.db.Collection(m.Table()).Find(db.Cond{"email": email}).One(&u)
	if err != nil {
		if errors.Is(err, db.ErrNoMoreRows) {
			return nil, ErrNoMoreRows
		}
		return nil, err
	}

	return &u, nil
}

func (m UserModel) Insert(u *User) error {
	newHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return err
	}
	u.Password = string(newHash)
	u.CreatedAt = time.Now()
	col := m.db.Collection(m.Table())
	res, err := col.Insert(u)
	if err != nil {
		switch {
		case errHasDuplicate(err, "user_email_key"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	u.ID = convertUpperToInt(res.ID())
	return nil
}

func (u *User) ComparePassword(plainPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (u UserModel) Authenticate(email, password string) (*User, error) {
	user, err := u.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if !user.Activated {
		return nil, ErrUserNotActive
	}
	match, err := user.ComparePassword(password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, ErrInvalidLogin
	}
	return user, nil

}
