package user

import "context"

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_user/user.go -source=./repository.go -package=mockuser

type UserRepository interface {
	Get(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, options ...SearchOption) (*User, error)
	Save(ctx context.Context, user *User) (*User, error)
}

type SearchParams struct {
	IDs []string
}

type SearchOption interface {
	Apply(*SearchParams)
}

type optionFunc func(*SearchParams)

func (f optionFunc) Apply(p *SearchParams) {
	f(p)
}

func WithIDs(ids []string) SearchOption {
	return optionFunc(func(p *SearchParams) {
		p.IDs = ids
	})
}
