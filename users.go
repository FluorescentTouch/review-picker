package main

import "sync"

type storage interface {
	Read(any) error
	Write(any) error
}

type Users struct {
	u map[string]User
	s storage
	l *sync.Mutex
}

type User struct {
	Name     string
	Vacation bool
}

func NewUsers(s storage) (Users, error) {
	u := Users{
		u: make(map[string]User),
		s: s,
		l: &sync.Mutex{},
	}
	err := s.Read(&u.u)
	return u, err
}

func (u Users) AddUser(user User) error {
	u.l.Lock()
	defer u.l.Unlock()

	u.u[user.Name] = user

	return u.s.Write(u.u)
}
