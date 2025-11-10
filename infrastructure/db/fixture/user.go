package fixture

import (
	guuid "github.com/google/uuid"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/uuid"
)

type userRow struct {
	id       guuid.UUID
	uid      string
	provider string
	email    *string
}

// NewUser prepares a user row based on the domain model. Call Setup to insert.
func (f *Fixture) NewUser(user *duser.User) {
	var id guuid.UUID
	if user.ID() == "" {
		id = uuid.NewV7()
	} else {
		parsed, err := guuid.Parse(user.ID())
		if err != nil {
			panic("invalid user ID: " + user.ID())
		}
		id = parsed
	}
	uid := user.UID()
	if uid == "" {
		panic("cannot add user with empty uid")
	}
	provider := user.Provider()
	if provider == "" {
		provider = duser.ProviderAnonymous
	}

	u := userRow{
		id:       id,
		uid:      uid,
		provider: string(provider),
		email:    user.Email(),
	}
	f.users = append(f.users, u)

	// Map alias to generated UUID
	f.idMap[uid] = id
}
