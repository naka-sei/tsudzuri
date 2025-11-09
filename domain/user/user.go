package user

type User struct {
	id       string
	uid      string
	provider Provider
	email    *string
}

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

// UID returns the user's UID.
func (u *User) UID() string {
	return u.uid
}

// Provider returns the user's provider.
func (u *User) Provider() Provider {
	return u.provider
}

// Email returns the user's email.
func (u *User) Email() *string {
	return u.email
}

type Users []*User

// Login logins the user with the given uid, provider, and email.
func (u *User) Login(provider string, email *string) error {
	if email == nil {
		return ErrNoSpecifiedEmail
	}

	p := Provider(provider)
	if err := p.isValid(); err != nil {
		return err
	}

	if u.provider != ProviderAnonymous {
		return ErrAlreadyLoggedIn(u.provider)
	}

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
