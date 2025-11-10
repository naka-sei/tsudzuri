package page

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/db/ent"
	entlinkitem "github.com/naka-sei/tsudzuri/infrastructure/db/ent/linkitem"
	entpage "github.com/naka-sei/tsudzuri/infrastructure/db/ent/page"

	"github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
)

type pageRepository struct {
	conn *postgres.Connection
}

func NewPageRepository(conn *postgres.Connection) dpage.PageRepository {
	return &pageRepository{conn: conn}
}

// Get fetches a page by ID with creator, links, and invited users.
func (r *pageRepository) Get(ctx context.Context, id string) (*dpage.Page, error) {
	if id == "" {
		return nil, nil
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid page id: %w", err)
	}

	client := r.conn.ReadOnlyDB(ctx)
	found, err := client.Page.
		Query().
		Where(entpage.IDEQ(uid)).
		WithCreator().
		WithLinkItems(func(q *ent.LinkItemQuery) { q.Order(ent.Asc(entlinkitem.FieldPriority)) }).
		WithInvitedUsers().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return r.entToDomain(found)
}

// List returns pages filtered by search options.
func (r *pageRepository) List(ctx context.Context, options ...dpage.SearchOption) ([]*dpage.Page, error) {
	params := dpage.SearchParams{}
	for _, opt := range options {
		opt.Apply(&params)
	}

	client := r.conn.ReadOnlyDB(ctx)
	q := client.Page.Query().
		WithCreator().
		WithLinkItems(func(q *ent.LinkItemQuery) { q.Order(ent.Asc(entlinkitem.FieldPriority)) }).
		WithInvitedUsers()

	if len(params.IDs) > 0 {
		uids := make([]uuid.UUID, 0, len(params.IDs))
		for _, id := range params.IDs {
			if parsed, err := uuid.Parse(id); err == nil {
				uids = append(uids, parsed)
			}
		}
		if len(uids) > 0 {
			q = q.Where(entpage.IDIn(uids...))
		}
	}
	if params.CreatedByUserID != "" {
		if cid, err := uuid.Parse(params.CreatedByUserID); err == nil {
			q = q.Where(entpage.CreatorIDEQ(cid))
		}
	}

	q = q.Order(ent.Asc(entpage.FieldID))
	page := postgres.PtrInt32ToInt(params.Page)
	pageSize := postgres.PtrInt32ToInt(params.PageSize)
	if pageSize > 0 {
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * pageSize
		q = q.Offset(offset).Limit(pageSize)
	}

	list, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	pages := make([]*dpage.Page, 0, len(list))
	for _, p := range list {
		dp, err := r.entToDomain(p)
		if err != nil {
			return nil, err
		}
		pages = append(pages, dp)
	}
	return pages, nil
}

// Save inserts or updates a page and its link items.
func (r *pageRepository) Save(ctx context.Context, pg *dpage.Page) (*dpage.Page, error) {
	if pg == nil {
		return nil, errors.New("nil page")
	}
	client := r.conn.WriteDB(ctx)

	// Upsert pattern: if ID empty -> create, else update title and sync links.
	var pageID uuid.UUID
	if pg.ID() == "" { // create
		creatorUUID, err := uuid.Parse(pg.CreatedBy().ID())
		if err != nil {
			return nil, fmt.Errorf("invalid creator id: %w", err)
		}
		created, err := client.Page.Create().
			SetTitle(pg.Title()).
			SetCreatorID(creatorUUID).
			SetInviteCode(pg.InviteCode()).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		// Hold created page ID for subsequent operations; reflect to domain later.
		pageID = created.ID
	} else { // update basic fields
		pid, err := uuid.Parse(pg.ID())
		if err != nil {
			return nil, fmt.Errorf("invalid page id: %w", err)
		}
		if _, err := client.Page.UpdateOneID(pid).SetTitle(pg.Title()).Save(ctx); err != nil {
			return nil, err
		}
		// Sync link items: simplistic approach delete then recreate in priority order.
		if _, err := client.LinkItem.Delete().Where(entlinkitem.PageIDEQ(pid)).Exec(ctx); err != nil {
			return nil, err
		}
		pageID = pid
	}

	// (Re)create link items from domain state
	if len(pg.Links()) > 0 {
		bulk := make([]*ent.LinkItemCreate, 0, len(pg.Links()))
		slices.SortFunc(pg.Links(), func(a, b dpage.Link) int { return a.Priority() - b.Priority() })
		for _, l := range pg.Links() {
			li := client.LinkItem.Create().
				SetPageID(pageID).
				SetURL(l.URL()).
				SetPriority(l.Priority())
			if m := l.Memo(); m != "" {
				memo := l.Memo()
				li.SetMemo(memo)
			}
			bulk = append(bulk, li)
		}
		if _, err := client.LinkItem.CreateBulk(bulk...).Save(ctx); err != nil {
			return nil, err
		}
	}

	return dpage.ReconstructPage(pageID.String(), pg.Title(), *pg.CreatedBy(), pg.InviteCode(), pg.Links(), pg.InvitedUsers()), nil
}

// DeleteByID deletes a page by ID (cascade relies on FK / DB constraints).
func (r *pageRepository) DeleteByID(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	pid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid page id: %w", err)
	}
	client := r.conn.WriteDB(ctx)
	return client.Page.DeleteOneID(pid).Exec(ctx)
}

// Helper: convert ent.Page to domain.Page.
func (r *pageRepository) entToDomain(p *ent.Page) (*dpage.Page, error) {
	if p == nil {
		return nil, nil
	}
	if p.Edges.Creator == nil {
		return nil, errors.New("page creator not loaded")
	}
	creator := r.entUserToDomain(p.Edges.Creator)
	links := make(dpage.Links, 0, len(p.Edges.LinkItems))
	for _, li := range p.Edges.LinkItems {
		memo := ""
		if li.Memo != nil {
			memo = *li.Memo
		}
		links = append(links, dpage.ReconstructLink(li.URL, memo, li.Priority))
	}
	invited := make(duser.Users, 0, len(p.Edges.InvitedUsers))
	for _, u := range p.Edges.InvitedUsers {
		invited = append(invited, r.entUserToDomain(u))
	}
	return dpage.ReconstructPage(p.ID.String(), p.Title, *creator, p.InviteCode, links, invited), nil
}

func (r *pageRepository) entUserToDomain(u *ent.User) *duser.User {
	if u == nil {
		return nil
	}
	return duser.ReconstructUser(u.ID.String(), u.UID, string(u.Provider), u.Email)
}
