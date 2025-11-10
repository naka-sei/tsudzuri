package page

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/infrastructure/db/fixture"
	"github.com/naka-sei/tsudzuri/infrastructure/db/postgres"
	"github.com/naka-sei/tsudzuri/pkg/ptr"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
)

func pageCmpOpts() []cmp.Option {
	return []cmp.Option{
		cmp.AllowUnexported(dpage.Link{}, dpage.Page{}, duser.User{}),
		cmpopts.SortSlices(func(a, b dpage.Link) bool { return a.Priority() < b.Priority() }),
		// Treat nil and empty slices as equal (e.g., invited users)
		cmpopts.EquateEmpty(),
	}
}

func TestPageRepository_Get(t *testing.T) {
	type args struct {
		id string
	}
	type want struct {
		page *dpage.Page
		err  error
	}

	tests := []struct {
		name    string
		prepare func(*fixture.Fixture)
		args    func(*fixture.Fixture) args
		want    func(*fixture.Fixture) want
	}{
		{
			name: "success",
			prepare: func(fx *fixture.Fixture) {
				// Use deterministic IDs in fixture and expectations to avoid alias resolution.
				creator := duser.ReconstructUser("", "creator-uid", string(duser.ProviderGoogle), ptr.Ptr("creator@example.com"))
				links := dpage.Links{
					dpage.ReconstructLink("https://example.com/1", "first memo", 1),
					dpage.ReconstructLink("https://example.com/2", "second memo", 2),
				}
				page := dpage.ReconstructPage("", "success", *creator, "INVGET01", links, nil)
				fx.NewUser(creator)
				fx.NewPage(page)
			},
			args: func(f *fixture.Fixture) args {
				return args{id: f.ID("success")}
			},
			want: func(f *fixture.Fixture) want {
				return want{
					page: dpage.ReconstructPage(
						f.ID("success"),
						"success",
						*duser.ReconstructUser(f.ID("creator-uid"), "creator-uid", string(duser.ProviderGoogle), ptr.Ptr("creator@example.com")),
						"INVGET01",
						dpage.Links{
							dpage.ReconstructLink("https://example.com/1", "first memo", 1),
							dpage.ReconstructLink("https://example.com/2", "second memo", 2),
						},
						nil,
					),
				}
			},
		},
		{
			name: "empty_id",
			args: func(f *fixture.Fixture) args { return args{id: ""} },
			want: func(f *fixture.Fixture) want { return want{page: nil} },
		},
		{
			name: "not_found",
			args: func(f *fixture.Fixture) args { return args{id: uuid.NewString()} },
			want: func(f *fixture.Fixture) want { return want{page: nil} },
		},
		{
			name: "invalid_id",
			args: func(f *fixture.Fixture) args { return args{id: "invalid"} },
			want: func(f *fixture.Fixture) want {
				return want{err: func() error {
					_, invalidUUIDErr := uuid.Parse("invalid")
					return invalidUUIDErr
				}()}
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn := postgres.SetupTestDBConnection(t)

			fx := fixture.New()
			if tt.prepare != nil {
				tt.prepare(fx)
			}

			if err := fx.Setup(ctx, t, conn.ReadOnlyDB(ctx)); err != nil {
				t.Fatalf("failed to setup fixture: %v", err)
			}

			args := tt.args(fx)
			want := tt.want(fx)

			repo := NewPageRepository(conn)
			got, err := repo.Get(ctx, args.id)
			testutil.EqualErr(t, want.err, err)

			if diff := cmp.Diff(want.page, got, pageCmpOpts()...); diff != "" {
				t.Fatalf("page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPageRepository_List tests listing pages with various filters.
func TestPageRepository_List(t *testing.T) {
	type args struct {
		opts []dpage.SearchOption
	}
	type want struct {
		pages []*dpage.Page
		err   error
	}

	tests := []struct {
		name    string
		prepare func(*fixture.Fixture)
		args    func(*fixture.Fixture) args
		want    func(*fixture.Fixture) want
	}{
		{
			name: "success_all",
			prepare: func(fx *fixture.Fixture) {
				creator := duser.ReconstructUser("", "creator-uid-1", string(duser.ProviderGoogle), ptr.Ptr("c1@example.com"))
				pageA := dpage.ReconstructPage("", "list-A", *creator, "INVLISTA", dpage.Links{
					dpage.ReconstructLink("https://example.com/a1", "a1", 1),
				}, nil)
				pageB := dpage.ReconstructPage("", "list-B", *creator, "INVLISTB", dpage.Links{
					dpage.ReconstructLink("https://example.com/b1", "b1", 1),
				}, nil)
				fx.NewUser(creator)
				fx.NewPage(pageA)
				fx.NewPage(pageB)
			},
			args: func(fx *fixture.Fixture) args { return args{} },
			want: func(fx *fixture.Fixture) want {
				creator := duser.ReconstructUser(fx.ID("creator-uid-1"), "creator-uid-1", string(duser.ProviderGoogle), ptr.Ptr("c1@example.com"))
				return want{pages: []*dpage.Page{
					dpage.ReconstructPage(fx.ID("list-A"), "list-A", *creator, "INVLISTA", dpage.Links{dpage.ReconstructLink("https://example.com/a1", "a1", 1)}, nil),
					dpage.ReconstructPage(fx.ID("list-B"), "list-B", *creator, "INVLISTB", dpage.Links{dpage.ReconstructLink("https://example.com/b1", "b1", 1)}, nil),
				}}
			},
		},
		{
			name: "filter_by_ids",
			prepare: func(fx *fixture.Fixture) {
				creator := duser.ReconstructUser("", "creator-uid-2", string(duser.ProviderGoogle), ptr.Ptr("c2@example.com"))
				pageA := dpage.ReconstructPage("", "list-C", *creator, "INVLISTC", nil, nil)
				pageB := dpage.ReconstructPage("", "list-D", *creator, "INVLISTD", nil, nil)
				fx.NewUser(creator)
				fx.NewPage(pageA)
				fx.NewPage(pageB)
			},
			args: func(fx *fixture.Fixture) args {
				return args{opts: []dpage.SearchOption{dpage.WithIDs([]string{fx.ID("list-C")})}}
			},
			want: func(fx *fixture.Fixture) want {
				creator := duser.ReconstructUser(fx.ID("creator-uid-2"), "creator-uid-2", string(duser.ProviderGoogle), ptr.Ptr("c2@example.com"))
				return want{pages: []*dpage.Page{
					dpage.ReconstructPage(fx.ID("list-C"), "list-C", *creator, "INVLISTC", nil, nil),
				}}
			},
		},
		{
			name: "filter_by_creator",
			prepare: func(fx *fixture.Fixture) {
				creator1 := duser.ReconstructUser("", "creator-uid-3a", string(duser.ProviderGoogle), ptr.Ptr("c3a@example.com"))
				creator2 := duser.ReconstructUser("", "creator-uid-3b", string(duser.ProviderGoogle), ptr.Ptr("c3b@example.com"))
				pageA := dpage.ReconstructPage("", "list-E", *creator1, "INVLISTE", nil, nil)
				pageB := dpage.ReconstructPage("", "list-F", *creator2, "INVLISTF", nil, nil)
				fx.NewUser(creator1)
				fx.NewUser(creator2)
				fx.NewPage(pageA)
				fx.NewPage(pageB)
			},
			args: func(fx *fixture.Fixture) args {
				return args{opts: []dpage.SearchOption{dpage.WithCreatedByUserID(fx.ID("creator-uid-3a"))}}
			},
			want: func(fx *fixture.Fixture) want {
				creator1 := duser.ReconstructUser(fx.ID("creator-uid-3a"), "creator-uid-3a", string(duser.ProviderGoogle), ptr.Ptr("c3a@example.com"))
				return want{pages: []*dpage.Page{dpage.ReconstructPage(fx.ID("list-E"), "list-E", *creator1, "INVLISTE", nil, nil)}}
			},
		},
		{
			name: "invalid_ids_ignored",
			prepare: func(fx *fixture.Fixture) {
				creator := duser.ReconstructUser("", "creator-uid-4", string(duser.ProviderGoogle), ptr.Ptr("c4@example.com"))
				page := dpage.ReconstructPage("", "list-G", *creator, "INVLISTG", nil, nil)
				fx.NewUser(creator)
				fx.NewPage(page)
			},
			args: func(fx *fixture.Fixture) args {
				return args{opts: []dpage.SearchOption{dpage.WithIDs([]string{"invalid", fx.ID("list-G")})}}
			},
			want: func(fx *fixture.Fixture) want {
				creator := duser.ReconstructUser(fx.ID("creator-uid-4"), "creator-uid-4", string(duser.ProviderGoogle), ptr.Ptr("c4@example.com"))
				return want{pages: []*dpage.Page{dpage.ReconstructPage(fx.ID("list-G"), "list-G", *creator, "INVLISTG", nil, nil)}}
			},
		},
		{
			name: "empty",
			args: func(fx *fixture.Fixture) args { return args{} },
			want: func(fx *fixture.Fixture) want { return want{} },
		},
	}

	ctx := context.Background()
	cmpOpts := append(pageCmpOpts(), cmpopts.SortSlices(func(a, b *dpage.Page) bool { return a.Title() < b.Title() }))

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conn := postgres.SetupTestDBConnection(t)
			fx := fixture.New()
			if tt.prepare != nil {
				tt.prepare(fx)
			}
			if err := fx.Setup(ctx, t, conn.ReadOnlyDB(ctx)); err != nil {
				t.Fatalf("failed to setup fixture: %v", err)
			}
			args := tt.args(fx)
			want := tt.want(fx)
			repo := NewPageRepository(conn)
			got, err := repo.List(ctx, args.opts...)
			testutil.EqualErr(t, want.err, err)
			if diff := cmp.Diff(want.pages, got, cmpOpts...); diff != "" {
				t.Fatalf("pages mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPageRepository_Save follows the same table style as others: name, prepare, args, want.
func TestPageRepository_Save(t *testing.T) {
	type args struct {
		page *dpage.Page
	}
	type want struct {
		page *dpage.Page
		err  error
	}

	tests := []struct {
		name    string
		prepare func(*fixture.Fixture)
		args    func(*fixture.Fixture) args
		want    func(*fixture.Fixture) want
	}{
		{
			name: "create_success",
			prepare: func(fx *fixture.Fixture) {
				// Prepare creator user with deterministic ID via alias mapping
				creator := duser.ReconstructUser("", "creator-save-uid", string(duser.ProviderGoogle), ptr.Ptr("save@example.com"))
				fx.NewUser(creator)
			},
			args: func(fx *fixture.Fixture) args {
				creator := duser.ReconstructUser(fx.ID("creator-save-uid"), "creator-save-uid", string(duser.ProviderGoogle), ptr.Ptr("save@example.com"))
				pg := dpage.ReconstructPage("", "save-create", *creator, "INVCR01", dpage.Links{
					dpage.ReconstructLink("https://create.com/1", "c1", 1),
					dpage.ReconstructLink("https://create.com/2", "c2", 2),
				}, nil)
				return args{page: pg}
			},
			want: func(fx *fixture.Fixture) want {
				creator := duser.ReconstructUser(fx.ID("creator-save-uid"), "creator-save-uid", string(duser.ProviderGoogle), ptr.Ptr("save@example.com"))
				// ID is generated at save time; leave it empty here and fill with got.ID in assertion
				expected := dpage.ReconstructPage("", "save-create", *creator, "INVCR01", dpage.Links{
					dpage.ReconstructLink("https://create.com/1", "c1", 1),
					dpage.ReconstructLink("https://create.com/2", "c2", 2),
				}, nil)
				return want{page: expected}
			},
		},
		{
			name: "update_success_title_and_links_edit",
			prepare: func(fx *fixture.Fixture) {
				creator := duser.ReconstructUser("", "creator-update-uid", string(duser.ProviderGoogle), ptr.Ptr("update@example.com"))
				original := dpage.ReconstructPage("", "save-update-original", *creator, "INVUP01", dpage.Links{
					dpage.ReconstructLink("https://update.com/1", "u1", 1),
					dpage.ReconstructLink("https://update.com/2", "u2", 2),
				}, nil)
				fx.NewUser(creator)
				fx.NewPage(original)
			},
			args: func(fx *fixture.Fixture) args {
				creator := duser.ReconstructUser(fx.ID("creator-update-uid"), "creator-update-uid", string(duser.ProviderGoogle), ptr.Ptr("update@example.com"))
				updated := dpage.ReconstructPage(fx.ID("save-update-original"), "save-update-new", *creator, "INVUP01", dpage.Links{
					dpage.ReconstructLink("https://update.com/2", "u2-new", 1),
					dpage.ReconstructLink("https://update.com/1", "u1-new", 2),
				}, nil)
				return args{page: updated}
			},
			want: func(fx *fixture.Fixture) want {
				creator := duser.ReconstructUser(fx.ID("creator-update-uid"), "creator-update-uid", string(duser.ProviderGoogle), ptr.Ptr("update@example.com"))
				expected := dpage.ReconstructPage(fx.ID("save-update-original"), "save-update-new", *creator, "INVUP01", dpage.Links{
					dpage.ReconstructLink("https://update.com/2", "u2-new", 1),
					dpage.ReconstructLink("https://update.com/1", "u1-new", 2),
				}, nil)
				return want{page: expected}
			},
		},
		{
			name: "create_invalid_creator_id",
			args: func(fx *fixture.Fixture) args {
				badCreator := duser.ReconstructUser("invalid", "creator-bad", string(duser.ProviderGoogle), ptr.Ptr("bad@example.com"))
				pg := dpage.ReconstructPage("", "save-invalid", *badCreator, "INVINVAL", nil, nil)
				return args{page: pg}
			},
			want: func(fx *fixture.Fixture) want {
				_, e := uuid.Parse("invalid")
				return want{err: e}
			},
		},
		{
			name: "update_invalid_page_id",
			prepare: func(fx *fixture.Fixture) {
				creator := duser.ReconstructUser("", "creator-up-bad", string(duser.ProviderGoogle), ptr.Ptr("upbad@example.com"))
				fx.NewUser(creator)
			},
			args: func(fx *fixture.Fixture) args {
				creator := duser.ReconstructUser(fx.ID("creator-up-bad"), "creator-up-bad", string(duser.ProviderGoogle), ptr.Ptr("upbad@example.com"))
				pg := dpage.ReconstructPage("invalid", "bad-update", *creator, "INVUPBAD", nil, nil)
				return args{page: pg}
			},
			want: func(fx *fixture.Fixture) want {
				_, e := uuid.Parse("invalid")
				return want{err: e}
			},
		},
		{
			name: "nil_page",
			args: func(fx *fixture.Fixture) args { return args{page: nil} },
			want: func(fx *fixture.Fixture) want { return want{err: errors.New("nil page")} },
		},
	}

	ctx := context.Background()
	cmpOpts := append(pageCmpOpts(), cmpopts.IgnoreFields(dpage.Page{}, "id"))
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conn := postgres.SetupTestDBConnection(t)
			fx := fixture.New()
			if tt.prepare != nil {
				tt.prepare(fx)
			}
			if err := fx.Setup(ctx, t, conn.ReadOnlyDB(ctx)); err != nil {
				t.Fatalf("failed to setup fixture: %v", err)
			}

			repo := NewPageRepository(conn)
			a := tt.args(fx)
			w := tt.want(fx)

			res, err := repo.Save(ctx, a.page)
			testutil.EqualErr(t, w.err, err)
			if w.err != nil {
				return
			}

			got, err := repo.Get(ctx, res.ID())
			if err != nil {
				t.Fatalf("failed to get saved page: %v", err)
			}

			if diff := cmp.Diff(w.page, got, cmpOpts...); diff != "" {
				t.Fatalf("saved page mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestPageRepository_DeleteByID tests deleting a page.
func TestPageRepository_DeleteByID(t *testing.T) {
	type args struct{ id string }
	type want struct {
		page *dpage.Page
		err  error
	}
	tests := []struct {
		name    string
		prepare func(*fixture.Fixture)
		args    func(*fixture.Fixture) args
		want    func(*fixture.Fixture) want
	}{
		{
			name: "success",
			prepare: func(fx *fixture.Fixture) {
				creator := duser.ReconstructUser("", "creator-del-uid", string(duser.ProviderGoogle), ptr.Ptr("del@example.com"))
				page := dpage.ReconstructPage("", "del-page", *creator, "INVDEL01", nil, nil)
				fx.NewUser(creator)
				fx.NewPage(page)
			},
			args: func(fx *fixture.Fixture) args { return args{id: fx.ID("del-page")} },
			want: func(fx *fixture.Fixture) want { return want{page: nil} },
		},
		{
			name: "empty_id",
			args: func(fx *fixture.Fixture) args { return args{id: ""} },
			want: func(fx *fixture.Fixture) want { return want{page: nil} },
		},
		{
			name: "invalid_id",
			args: func(fx *fixture.Fixture) args { return args{id: "invalid"} },
			want: func(fx *fixture.Fixture) want {
				return want{err: func() error { _, e := uuid.Parse("invalid"); return e }()}
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conn := postgres.SetupTestDBConnection(t)
			fx := fixture.New()
			if tt.prepare != nil {
				tt.prepare(fx)
			}
			if err := fx.Setup(ctx, t, conn.ReadOnlyDB(ctx)); err != nil {
				t.Fatalf("failed to setup fixture: %v", err)
			}
			repo := NewPageRepository(conn)
			a := tt.args(fx)
			w := tt.want(fx)
			err := repo.DeleteByID(ctx, a.id)
			testutil.EqualErr(t, w.err, err)
			if w.err != nil {
				return
			}
			// Verify by fetching the page; should be nil when deletion succeeds or id is empty.
			got, gErr := repo.Get(ctx, a.id)
			if gErr != nil {
				t.Fatalf("Get after delete failed: %v", gErr)
			}
			if diff := cmp.Diff(w.page, got, pageCmpOpts()...); diff != "" {
				t.Fatalf("page after delete mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
