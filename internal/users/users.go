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

	// map reads has poor versatility in low dimensional sets, so transform to rand read from slice instead
	trimmed := make(map[string]User)
	for k, v := range u.u[chatID] {
		if k == except || v.Vacation {
			continue
		}
		trimmed[k] = v
	}

	if len(trimmed) == 0 {
		return User{}, ErrNotFound
	}

	list := make([]User, 0, len(trimmed))
	for _, v := range trimmed {
		list = append(list, v)
	}

	index := rand.Intn(len(list) - 1)
	return list[index], nil
}

func (u Users) Rand2(chatID int64, except string) ([2]User, error) {
	u.l.Lock()
	defer u.l.Unlock()

	// map reads has poor versatility in low dimensional sets, so transform to rand read from slice instead
	trimmed := make(map[string]User)
	for k, v := range u.u[chatID] {
		if k == except || v.Vacation {
			continue
		}
		trimmed[k] = v
	}

	if len(trimmed) < 2 {
		return [2]User{}, ErrNotFound
	}

	list := make([]User, 0, len(trimmed))
	for _, v := range trimmed {
		list = append(list, v)
	}

	candidates := make(map[User]struct{}, 2)
	for len(candidates) != 2 {
		index := rand.Intn(len(list) - 1)
		candidates[list[index]] = struct{}{}
	}

	out := make([]User, 0, 2)
	for k := range candidates {
		out = append(out, k)
	}

	return [2]User{out[0], out[1]}, nil
}
