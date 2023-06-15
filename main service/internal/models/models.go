package models

type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	Activated bool   `json:"activated"`
	Roles     string `json:"roles"`
}

var AnonymousUser = &User{}
