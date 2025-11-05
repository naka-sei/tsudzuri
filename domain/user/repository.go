package user

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination mock/mock_user/user.go -source=./repository.go -package=mockuser

type UserRepository interface {
	FindByID(id int64) (*User, error)
	Save(user *User) error
}
