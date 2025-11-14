package fixture

import (
	guuid "github.com/google/uuid"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	"github.com/naka-sei/tsudzuri/pkg/ptr"
	"github.com/naka-sei/tsudzuri/pkg/uuid"
)

type pageRow struct {
	id        guuid.UUID
	title     string
	creatorID guuid.UUID
	invite    string
}

type linkItemRow struct {
	pageID   guuid.UUID
	url      string
	memo     *string
	priority int
}

// NewPage prepares a page row based on the domain model. Call Setup to insert.
func (f *Fixture) NewPage(page *dpage.Page) {
	var pageID guuid.UUID
	if page.ID() == "" {
		pageID = uuid.NewV7()
	} else {
		parsed, err := guuid.Parse(page.ID())
		if err != nil {
			panic("invalid page ID: " + page.ID())
		}
		pageID = parsed
	}
	title := page.Title()
	if title == "" {
		panic("cannot add page with empty title")
	}
	var creatorID guuid.UUID
	if page.CreatedBy().ID() == "" {
		uid := page.CreatedBy().UID()
		mappedID, ok := f.idMap[uid]
		if !ok {
			panic("cannot add page for user not in fixture: " + uid)
		}
		creatorID = mappedID
	} else {
		parsed, err := guuid.Parse(page.CreatedBy().ID())
		if err != nil {
			panic("invalid creator ID: " + page.CreatedBy().ID())
		}
		creatorID = parsed
	}
	f.pages = append(f.pages, pageRow{
		id:        pageID,
		title:     title,
		creatorID: creatorID,
		invite:    page.InviteCode(),
	})

	// Map alias to generated UUID
	f.idMap[title] = pageID

	f.linkItems = append(f.linkItems, func() []linkItemRow {
		rows := make([]linkItemRow, 0, len(page.Links()))
		for i, l := range page.Links() {
			url := l.URL()
			if url == "" {
				panic("cannot add link item with empty URL")
			}
			rows = append(rows, linkItemRow{
				pageID:   pageID,
				url:      url,
				memo:     ptr.Ptr(l.Memo()),
				priority: i + 1,
			})
		}
		return rows
	}()...)

	for _, invited := range page.InvitedUsers() {
		if invited == nil {
			panic("cannot add nil invited user")
		}
		key := invited.ID()
		if key == "" {
			key = invited.UID()
		}
		if key == "" {
			panic("invited user must have id or uid")
		}
		f.pageUsers = append(f.pageUsers, pageUserRow{pageKey: title, userKey: key})
	}
}
