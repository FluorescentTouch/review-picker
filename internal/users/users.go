package users

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not enough active users: absent or vacationing")
)

type userStorage interface {
	Read(any) error
	Write(any) error
}

type Users struct {
	u map[int64]map[string]User // chatID:userID:User
	s userStorage
	l *sync.Mutex
}

type User struct {
	Name     string
	Vacation bool
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewUsers(s userStorage) (Users, error) {
	u := Users{
		u: make(map[int64]map[string]User),
		s: s,
		l: &sync.Mutex{},
	}
	err := s.Read(&u.u)
	return u, err
}

func (u Users) AddUser(chatID int64, user User) error {
	u.l.Lock()
	defer u.l.Unlock()

	if _, ok := u.u[chatID]; ok {
		u.u[chatID][user.Name] = user
	} else {
		u.u[chatID] = map[string]User{user.Name: user}
	}

	return u.s.Write(u.u)
}

func (u Users) Rand(chatID int64, except string) (User, error) {
	u.l.Lock()
	defer u.l.Unlock()

	for name, user := range u.u[chatID] {
		if name != except && !user.Vacation {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

func (u Users) Rand2(chatID int64, except string) ([2]User, error) {
	u.l.Lock()
	defer u.l.Unlock()

	index := 0
	users := [2]User{}
	for name, user := range u.u[chatID] {
		if name != except && !user.Vacation {
			users[index] = user
			index++
			if index > 1 {
				return users, nil
			}
		}
	}

	return [2]User{}, ErrNotFound
}
