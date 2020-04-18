package models

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
)

var SecretSalt = "secret salt"

type UserPassword [sha512.Size]byte
type UserId uint64

const UserAuthKeySize = 1 + 8 + sha512.Size

type UserAuthKey [UserAuthKeySize]byte

const authKeyVersion = 1

type User struct {
	tableName     struct{}     `pg:"users"`
	ID            UserId       `pg:",pk"  json:"user_id"`
	CreatedAt     time.Time    `pg:"default:now()" json:"-"`
	UpdatedAt     time.Time    `pg:"default:now()" json:"-"`
	DeletedAt     *time.Time   `pg:",soft_delete" json:"deleted_at,omitempty"`
	Email         string       `pg:"type:varchar(500),unique,notnull" json:"email"`
	Password      UserPassword `json:"-"`
	PasswordPlain string       `json:"password_plain,omitempty" pg:"-"`
	Name          string       `pg:"name"  json:"name"`
}

func (u *User) SetPassword(password string) {
	u.Password = u.hash(password)
}

func (u User) hash(password string) UserPassword {
	return sha512.Sum512(
		append(
			append(
				[]byte(password), []byte(u.Email)...,
			),
			[]byte(SecretSalt)...,
		),
	)
}

func (u User) GetAuthKey() string {
	h := u.authKey()
	return base64.URLEncoding.EncodeToString(h[:])
}

func (u User) authKey() UserAuthKey {
	h := UserAuthKey{authKeyVersion, 0}
	binary.BigEndian.PutUint64(h[1:9], uint64(u.ID))
	z := u.AuthHash()
	copy(h[9:], z[:])
	return h
}

func (u User) AuthHash() UserPassword {
	z := sha512.Sum512(append(u.Password[:], []byte(SecretSalt)...))
	return z
}

func AuthKeyToUserSearch(k string) (*User, error) {
	t, err := base64.URLEncoding.DecodeString(k)
	if err != nil {
		return nil, err
	}
	if len(t) == 0 {
		return nil, fmt.Errorf("wrong length %d", len(t))
	}
	if t[0] != authKeyVersion || len(t) != UserAuthKeySize {
		return nil, fmt.Errorf("wrong version %d or length %d", t[0], len(t))
	}

	id := binary.BigEndian.Uint64(t[1:9])

	u := &User{}
	u.ID = UserId(id)
	copy(u.Password[:], t[9:])
	return u, nil
}
