package user

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

func userCmpOpts() []cmp.Option {
	return []cmp.Option{cmp.AllowUnexported(duser.User{}), cmpopts.EquateEmpty()}
}

func TestUserRepository_Get(t *testing.T) {
	type args struct{ id string }
	type want struct {
		user *duser.User
		err  error
	}
	tests := []struct {
		name    string
		prepare func(*fixture.Fixture)
		args    args
		want    func(*fixture.Fixture) want
	}{
		{
			name: "success",
			prepare: func(fx *fixture.Fixture) {
				u := duser.ReconstructUser("", "uid-get", string(duser.ProviderGoogle), ptr.Ptr("g@example.com"))
				fx.NewUser(u)
			},
			args: args{id: "uid-get"},
			want: func(fx *fixture.Fixture) want {
				return want{user: duser.ReconstructUser(fx.ID("uid-get"), "uid-get", string(duser.ProviderGoogle), ptr.Ptr("g@example.com"))}
			},
		},
		{
			name: "success_with_joined_pages",
			prepare: func(fx *fixture.Fixture) {
				invited := duser.ReconstructUser("", "uid-join", string(duser.ProviderGoogle), ptr.Ptr("join@example.com"))
				fx.NewUser(invited)
				creator := duser.ReconstructUser("", "creator", string(duser.ProviderGoogle), ptr.Ptr("creator@example.com"))
				fx.NewUser(creator)
				page := dpage.ReconstructPage("", "page-join", *creator, "joincode", nil, nil)
				fx.NewPage(page)
				fx.AddPageUser("page-join", "uid-join")
			},
			args: args{id: "uid-join"},
			want: func(fx *fixture.Fixture) want {
				return want{
					user: duser.ReconstructUser(
						fx.ID("uid-join"),
						"uid-join",
						string(duser.ProviderGoogle),
						ptr.Ptr("join@example.com"),
						duser.WithJoinedPageIDs([]string{fx.ID("page-join")}),
					),
				}
			},
		},
		{name: "empty_id", args: args{id: ""}, want: func(fx *fixture.Fixture) want { return want{} }},
		{name: "not_found", args: args{id: uuid.NewString()}, want: func(fx *fixture.Fixture) want { return want{} }},
		{name: "invalid_id", args: args{id: "invalid"}, want: func(fx *fixture.Fixture) want { return want{} }},
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
				t.Fatalf("fixture setup: %v", err)
			}
			repo := NewUserRepository(conn)
			a := tt.args
			w := tt.want(fx)
			got, err := repo.Get(ctx, a.id)
			testutil.EqualErr(t, w.err, err)
			if diff := cmp.Diff(w.user, got, userCmpOpts()...); diff != "" {
				t.Fatalf("user mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	type args struct{ opts []duser.SearchOption }
	type want struct {
		user *duser.User
		err  error
	}
	tests := []struct {
		name    string
		prepare func(*fixture.Fixture)
		args    func(*fixture.Fixture) args
		want    func(*fixture.Fixture) want
	}{
		{
			name: "success_first_match",
			prepare: func(fx *fixture.Fixture) {
				fx.NewUser(duser.ReconstructUser("", "uid-list-1", string(duser.ProviderGoogle), ptr.Ptr("l1@example.com")))
				fx.NewUser(duser.ReconstructUser("", "uid-list-2", string(duser.ProviderGoogle), ptr.Ptr("l2@example.com")))
			},
			args: func(fx *fixture.Fixture) args {
				return args{opts: []duser.SearchOption{duser.WithIDs([]string{fx.ID("uid-list-2"), fx.ID("uid-list-1")})}}
			},
			want: func(fx *fixture.Fixture) want {
				return want{user: duser.ReconstructUser(fx.ID("uid-list-2"), "uid-list-2", string(duser.ProviderGoogle), ptr.Ptr("l2@example.com"))}
			},
		},
		{
			name: "success_with_joined_pages",
			prepare: func(fx *fixture.Fixture) {
				invited := duser.ReconstructUser("", "uid-list-join", string(duser.ProviderGoogle), ptr.Ptr("join@example.com"))
				fx.NewUser(invited)
				creator := duser.ReconstructUser("", "creator-list", string(duser.ProviderGoogle), ptr.Ptr("creator@example.com"))
				fx.NewUser(creator)
				page := dpage.ReconstructPage("", "page-list-join", *creator, "listcode", nil, nil)
				fx.NewPage(page)
				fx.AddPageUser("page-list-join", "uid-list-join")
			},
			args: func(fx *fixture.Fixture) args {
				return args{opts: []duser.SearchOption{duser.WithIDs([]string{fx.ID("uid-list-join")})}}
			},
			want: func(fx *fixture.Fixture) want {
				return want{user: duser.ReconstructUser(
					fx.ID("uid-list-join"),
					"uid-list-join",
					string(duser.ProviderGoogle),
					ptr.Ptr("join@example.com"),
					duser.WithJoinedPageIDs([]string{fx.ID("page-list-join")}),
				)}
			},
		},
		{
			name: "invalid_ids_ignored",
			prepare: func(fx *fixture.Fixture) {
				fx.NewUser(duser.ReconstructUser("", "uid-list-3", string(duser.ProviderGoogle), ptr.Ptr("l3@example.com")))
			},
			args: func(fx *fixture.Fixture) args {
				return args{opts: []duser.SearchOption{duser.WithIDs([]string{"bad", fx.ID("uid-list-3")})}}
			},
			want: func(fx *fixture.Fixture) want {
				return want{user: duser.ReconstructUser(fx.ID("uid-list-3"), "uid-list-3", string(duser.ProviderGoogle), ptr.Ptr("l3@example.com"))}
			},
		},
		{name: "empty", args: func(fx *fixture.Fixture) args { return args{} }, want: func(fx *fixture.Fixture) want { return want{} }},
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
				t.Fatalf("fixture setup: %v", err)
			}
			repo := NewUserRepository(conn)
			a := tt.args(fx)
			w := tt.want(fx)
			got, err := repo.List(ctx, a.opts...)
			testutil.EqualErr(t, w.err, err)
			if diff := cmp.Diff(w.user, got, userCmpOpts()...); diff != "" {
				t.Fatalf("user mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUserRepository_Save(t *testing.T) {
	type args struct{ user *duser.User }
	type want struct {
		user *duser.User
		err  error
	}

	cmpOpts := append(userCmpOpts(), cmpopts.IgnoreFields(duser.User{}, "id"))
	tests := []struct {
		name    string
		prepare func(*fixture.Fixture)
		args    func(*fixture.Fixture) args
		want    func(*fixture.Fixture) want
	}{
		{
			name: "create_success",
			args: func(fx *fixture.Fixture) args { return args{user: duser.NewUser("uid-save-1")} },
			want: func(fx *fixture.Fixture) want {
				return want{user: duser.ReconstructUser("", "uid-save-1", string(duser.ProviderAnonymous), nil)}
			},
		},
		{
			name: "update_success",
			prepare: func(fx *fixture.Fixture) {
				fx.NewUser(duser.ReconstructUser("", "uid-save-2", string(duser.ProviderGoogle), ptr.Ptr("before@example.com")))
			},
			args: func(fx *fixture.Fixture) args {
				return args{user: duser.ReconstructUser(fx.ID("uid-save-2"), "uid-save-2", string(duser.ProviderFacebook), ptr.Ptr("after@example.com"))}
			},
			want: func(fx *fixture.Fixture) want {
				return want{user: duser.ReconstructUser(fx.ID("uid-save-2"), "uid-save-2", string(duser.ProviderFacebook), ptr.Ptr("after@example.com"))}
			},
		},
		{
			name: "invalid_id_update",
			args: func(fx *fixture.Fixture) args {
				return args{user: duser.ReconstructUser("bad", "uid-save-3", string(duser.ProviderGoogle), ptr.Ptr("x@example.com"))}
			},
			want: func(fx *fixture.Fixture) want {
				return want{err: func() error { _, e := uuid.Parse("bad"); return e }()}
			},
		},
		{name: "nil_user", args: func(fx *fixture.Fixture) args { return args{user: nil} }, want: func(fx *fixture.Fixture) want { return want{err: errors.New("nil user")} }},
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
				t.Fatalf("fixture setup: %v", err)
			}
			repo := NewUserRepository(conn)
			a := tt.args(fx)
			w := tt.want(fx)
			res, err := repo.Save(ctx, a.user)
			testutil.EqualErr(t, w.err, err)
			if w.err != nil {
				return
			}
			got, err := repo.Get(ctx, res.UID())
			if err != nil {
				t.Fatalf("Get after save failed: %v", err)
			}
			if diff := cmp.Diff(w.user, got, cmpOpts...); diff != "" {
				t.Fatalf("saved user mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
