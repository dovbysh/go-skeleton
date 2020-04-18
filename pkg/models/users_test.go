package models

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUser_GetAuthKey(t *testing.T) {
	u := User{
		ID:            2,
		CreatedAt:     time.Now(),
		Email:         "testEmail",
		PasswordPlain: "testPassword",
		Name:          "testName",
	}
	u.SetPassword(u.PasswordPlain)
	ah := u.AuthHash()
	t.Log(ah)
	bk := u.authKey()
	t.Log(bk)
	k := u.GetAuthKey()
	t.Log(k)

	userSearch, err := AuthKeyToUserSearch(base64.URLEncoding.EncodeToString(nil))
	assert.Error(t, err)
	assert.Empty(t, userSearch)

	userSearch, err = AuthKeyToUserSearch(k)
	assert.NoError(t, err)
	assert.NotEmpty(t, userSearch)
	assert.Equal(t, ah, userSearch.Password)
	// We will found user by id at database and check by AuthHash
}
