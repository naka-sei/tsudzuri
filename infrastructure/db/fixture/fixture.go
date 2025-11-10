package fixture

import (
	"context"
	"testing"

	guuid "github.com/google/uuid"

	"github.com/naka-sei/tsudzuri/infrastructure/db/ent"
	entuser "github.com/naka-sei/tsudzuri/infrastructure/db/ent/user"
)

// Fixture collects data to be inserted (prepared by New*), and Setup actually inserts them.
type Fixture struct {
	users     []userRow
	pages     []pageRow
	linkItems []linkItemRow

	// alias to generated UUID mapping (after Setup)
	idMap map[string]guuid.UUID
}

// New creates an empty Fixture.
func New() *Fixture { return &Fixture{idMap: map[string]guuid.UUID{}} }

func (f *Fixture) Setup(ctx context.Context, t *testing.T, client *ent.Client) error {
	t.Helper()

	// Insert users
	if len(f.users) > 0 {
		builders := make([]*ent.UserCreate, 0, len(f.users))
		for _, u := range f.users {
			b := client.User.Create().
				SetUID(u.uid).
				SetProvider(entuser.Provider(u.provider)).
				SetID(u.id)
			if u.email != nil {
				b = b.SetNillableEmail(u.email)
			}
			builders = append(builders, b)
		}
		_, err := client.User.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return err
		}
	}

	// Insert pages
	if len(f.pages) > 0 {
		builders := make([]*ent.PageCreate, 0, len(f.pages))
		for _, p := range f.pages {
			b := client.Page.Create().
				SetTitle(p.title).
				SetCreatorID(p.creatorID).
				SetInviteCode(p.invite).
				SetID(p.id)
			builders = append(builders, b)
		}
		_, err := client.Page.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return err
		}
	}

	// Insert link items after pages (need page IDs)
	if len(f.linkItems) > 0 {
		builders := make([]*ent.LinkItemCreate, 0, len(f.linkItems))
		for _, li := range f.linkItems {
			b := client.LinkItem.Create().
				SetPageID(li.pageID).
				SetURL(li.url).
				SetPriority(li.priority)
			if li.memo != nil {
				b = b.SetMemo(*li.memo)
			}
			builders = append(builders, b)
		}
		_, err := client.LinkItem.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// ID returns the real UUID corresponding to an alias or the ID itself if already real.
func (f *Fixture) ID(aliasOrID string) string {
	return f.resolveID(aliasOrID)
}

func (f *Fixture) resolveID(k string) string {
	if k == "" {
		return ""
	}
	if v, ok := f.idMap[k]; ok {
		return v.String()
	}
	return k // assume already a UUID
}
