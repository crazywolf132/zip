package git

import "strings"

type User struct {
	Name  string
	Email string
}

func (r *Repo) User() (*User, error) {
	nameCommand, err := r.Git("config", "--get", "user.name")
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(string(nameCommand))
	emailCommand, err := r.Git("config", "--get", "user.email")
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(string(emailCommand))
	return &User{
		Name:  name,
		Email: email,
	}, nil
}
