package page

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	mockpage "github.com/naka-sei/tsudzuri/domain/page/mock/mock_page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mocktxn "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
)

func TestEditUsecase_Edit(t *testing.T) {
	type mocks struct {
		pageRepo *mockpage.MockPageRepository
		txn      *mocktxn.MockTransactionService
	}
	type args struct {
		ctx    context.Context
		pageID string
		title  string
		links  dpage.Links
	}
	type want struct {
		err error
	}

	user := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)
	other := duser.ReconstructUser("user-id-2", "uid-2", "anonymous", nil)
	invitedUser := duser.ReconstructUser("user-id-3", "uid-3", "invited", nil)

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success_by_creator",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("page-1", "t1", *user, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)
				m.pageRepo.EXPECT().Save(gomock.Any(), page).Return(nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
				title:  "new title",
				links:  dpage.Links{},
			},
			want: want{err: nil},
		},
		{
			name: "success_by_invited_user",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("page-1", "t1", *user, "invite", dpage.Links{}, duser.Users{invitedUser})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)
				m.pageRepo.EXPECT().Save(gomock.Any(), page).Return(nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), invitedUser),
				pageID: "page-1",
				title:  "new title",
				links:  dpage.Links{},
			},
			want: want{err: nil},
		},
		{
			name: "repo_get_error",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(nil, errors.New("get error"))
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
				title:  "new title",
				links:  dpage.Links{},
			},
			want: want{err: errors.New("get error")},
		},
		{
			name: "page_not_found",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(nil, nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
				title:  "new title",
				links:  dpage.Links{},
			},
			want: want{err: ErrPageNotFound},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("page-1", "t1", *user, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(page, nil)
			},
			args: args{
				ctx:    context.Background(),
				pageID: "page-1",
				title:  "new title",
				links:  dpage.Links{},
			},
			want: want{err: duser.ErrUserNotFound},
		},
		{
			name: "unauthorized",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("page-2", "t2", *other, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-2").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-2",
				title:  "new title",
				links:  dpage.Links{},
			},
			want: want{err: dpage.ErrNotCreatedByUser},
		},
		{
			name: "save_error",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("page-1", "t1", *user, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)
				m.pageRepo.EXPECT().Save(gomock.Any(), page).Return(errors.New("save error"))
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
				title:  "new title",
				links:  dpage.Links{},
			},
			want: want{err: errors.New("save error")},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &mocks{
				pageRepo: mockpage.NewMockPageRepository(ctrl),
				txn:      mocktxn.NewMockTransactionService(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			u := NewEditUsecase(m.pageRepo, m.txn)
			err := u.Edit(tt.args.ctx, tt.args.pageID, tt.args.title, tt.args.links)
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
