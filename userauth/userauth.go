package userauth

import (

	"github.com/BurntSushi/toml"
	oauth2 "google.golang.org/api/oauth2/v2"
)

type UserAuth struct {
	UserList []string
}

func NewUserAuth(path string) *UserAuth {
	var config UserAuth
	toml.DecodeFile(path, &config)
	
	return &config
}

func (l *UserAuth) Authentication(token *oauth2.Tokeninfo) bool {
	for _, v := range l.UserList {
		if v == token.Email {
			return true
		}
	}
	return false
}
