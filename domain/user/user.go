package user

type User struct {
	id       string
	uid      string
	provider Provider
	email    *string
}

type Users []*User

// NewUser creates a new User instance.
func NewUser(uid string) *User {
	return &User{
		uid:      uid,
		provider: ProviderAnonymous,
	}
}

// ID returns the user's ID.
func (u *User) ID() string {
	return u.id
}

// Login logins the user with the given uid, provider, and email.
func (u *User) Login(uid string, provider string, email *string) error {
	if email == nil {
		return ErrNoSpecifiedEmail
	}

	if u.uid != uid {
		return ErrInvalidUID(uid)
	}

	p := Provider(provider)
	if err := p.isValid(); err != nil {
		return err
	}

	if u.provider != ProviderAnonymous {
		return ErrAlreadyLoggedIn(u.provider)
	}

	u.uid = uid
	u.provider = p
	u.email = email

	return nil
}

// ReconstructUser reconstructs a User instance from existing data.
func ReconstructUser(id string, uid string, provider string, email *string) *User {
	return &User{
		id:       id,
		uid:      uid,
		provider: Provider(provider),
		email:    email,
	}
}
