package user

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	mockuser "github.com/naka-sei/tsudzuri/domain/user/mock/mock_user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	"go.uber.org/mock/gomock"
)

func TestGetUsecase_Get(t *testing.T) {
	type fields struct {
		userRepo *mockuser.MockUserRepository
	}
	type args struct {
		ctx context.Context
	}
	type want struct {
		user *duser.User
		err  error
	}

	tests := []struct {
		name  string
		setup func(f *fields)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(f *fields) {
				expected := duser.ReconstructUser("id-1", "uid-1", string(duser.ProviderGoogle), nil, duser.WithJoinedPageIDs([]string{"page-1"}))
				f.userRepo.EXPECT().Get(gomock.Any(), "uid-1").Return(expected, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), duser.ReconstructUser("id-ctx", "uid-1", string(duser.ProviderGoogle), nil)),
			},
			want: want{
				user: duser.ReconstructUser("id-1", "uid-1", string(duser.ProviderGoogle), nil, duser.WithJoinedPageIDs([]string{"page-1"})),
			},
		},
		{
			name: "repository_returns_nil",
			setup: func(f *fields) {
				f.userRepo.EXPECT().Get(gomock.Any(), "uid-2").Return(nil, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), duser.ReconstructUser("id-ctx", "uid-2", string(duser.ProviderGoogle), nil)),
			},
			want: want{
				err: duser.ErrUserNotFound,
			},
		},
		{
			name: "repository_error",
			setup: func(f *fields) {
				f.userRepo.EXPECT().Get(gomock.Any(), "uid-3").Return(nil, errors.New("repo error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), duser.ReconstructUser("id-ctx", "uid-3", string(duser.ProviderGoogle), nil)),
			},
			want: want{
				err: errors.New("repo error"),
			},
		},
		{
			name:  "no_user_in_context",
			setup: func(f *fields) {},
			args:  args{ctx: context.Background()},
			want:  want{err: duser.ErrUserNotFound},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{userRepo: mockuser.NewMockUserRepository(ctrl)}
			if tt.setup != nil {
				tt.setup(f)
			}

			u := NewGetUsecase(f.userRepo)
			got, err := u.Get(tt.args.ctx)

			if diff := cmp.Diff(tt.want.user, got, cmp.AllowUnexported(duser.User{})); diff != "" {
				t.Fatalf("user mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
