package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/db/ent"
	entuser "github.com/naka-sei/tsudzuri/infrastructure/db/ent/user"
	"github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
)

type userRepository struct {
	conn *postgres.Connection
}

// NewUserRepository creates a new user repository backed by postgres/ent.
func NewUserRepository(conn *postgres.Connection) duser.UserRepository {
	return &userRepository{conn: conn}
}

// Get fetches a user by UID. Returns (nil, nil) if not found or id empty.
func (r *userRepository) Get(ctx context.Context, uid string) (*duser.User, error) {
	if uid == "" {
		return nil, nil
	}
	client := r.conn.ReadOnlyDB(ctx)
	found, err := client.User.Query().Where(entuser.UIDEQ(uid)).WithInvitedPages().Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return r.entToDomain(found), nil
}

// List returns users matching search options. Currently supports filtering by IDs.
// NOTE: Domain interface returns a single *User; keeping signature but implementation
// logically returns the first matched user if any. If multiple IDs provided, the first existing is returned.
func (r *userRepository) List(ctx context.Context, options ...duser.SearchOption) (*duser.User, error) {
	params := duser.SearchParams{}
	for _, opt := range options {
		opt.Apply(&params)
	}

	client := r.conn.ReadOnlyDB(ctx)

	// If IDs are specified, return the first existing user in the provided order.
	if len(params.IDs) > 0 {
		for _, id := range params.IDs {
			uid, err := uuid.Parse(id)
			if err != nil { // ignore invalid id
				continue
			}
			u, err := client.User.Query().Where(entuser.IDEQ(uid)).WithInvitedPages().Only(ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					continue
				}
				return nil, err
			}
			return r.entToDomain(u), nil
		}
		return nil, nil
	}

	// No filters -> align with tests and return no result.
	return nil, nil
}

// Save creates or updates a user.
// Create when user.ID() == ""; update provider/email when existing.
func (r *userRepository) Save(ctx context.Context, user *duser.User) (*duser.User, error) {
	if user == nil {
		return nil, errors.New("nil user")
	}
	client := r.conn.WriteDB(ctx)
	if user.ID() == "" { // create
		created, err := client.User.Create().
			SetUID(user.UID()).
			SetProvider(entuser.Provider(user.Provider())).
			SetNillableEmail(user.Email()).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		r.setUserID(user, created.ID.String())
		return user, nil
	}
	// update
	uid, err := uuid.Parse(user.ID())
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	upd := client.User.UpdateOneID(uid).
		SetUID(user.UID()).
		SetProvider(entuser.Provider(user.Provider())).
		SetNillableEmail(user.Email())
	if _, err := upd.Save(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

// entToDomain converts ent.User to domain.User.
func (r *userRepository) entToDomain(u *ent.User) *duser.User {
	if u == nil {
		return nil
	}
	var joinedIDs []string
	if pages, err := u.Edges.InvitedPagesOrErr(); err == nil {
		joinedIDs = make([]string, 0, len(pages))
		for _, p := range pages {
			joinedIDs = append(joinedIDs, p.ID.String())
		}
	}
	return duser.ReconstructUser(u.ID.String(), u.UID, string(u.Provider), u.Email, duser.WithJoinedPageIDs(joinedIDs))
}

// setUserID reconstructs user with new ID.
func (r *userRepository) setUserID(u *duser.User, id string) {
	*u = *duser.ReconstructUser(id, u.UID(), string(u.Provider()), u.Email(), duser.WithJoinedPageIDs(u.JoinedPageIDs()))
}
