package page

import (
	"context"
	"errors"
	"testing"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	mockpage "github.com/naka-sei/tsudzuri/domain/page/mock/mock_page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cmperr"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	mocktxn "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
	"go.uber.org/mock/gomock"
)

func TestLinkRemoveUsecase_LinkRemove(t *testing.T) {
	type mocks struct {
		pageRepo *mockpage.MockPageRepository
		txn      *mocktxn.MockTransactionService
	}
	type args struct {
		ctx   context.Context
		input LinkRemoveUsecaseInput
	}
	type want struct {
		err error
	}

	creator := duser.ReconstructUser("user-id-1", "creator-uid-1", "anonymous", nil)
	invitedUser := duser.ReconstructUser("user-id-2", "invited-uid-2", "anonymous", nil)
	unauthorizedUser := duser.ReconstructUser("user-id-3", "unauthorized-uid-3", "anonymous", nil)

	initialLinks := dpage.Links{
		dpage.ReconstructLink("https://link1.com", "Memo 1", 1),
		dpage.ReconstructLink("https://link2.com", "Memo 2", 2),
		dpage.ReconstructLink("https://link3.com", "Memo 3", 3),
	}
	invitedUsers := duser.Users{invitedUser}
	initialPage := dpage.ReconstructPage("page-1", "Test Page", *creator, "invite-code", initialLinks, invitedUsers)

	expectedLinks := dpage.Links{
		dpage.ReconstructLink("https://link1.com", "Memo 1", 1),
		dpage.ReconstructLink("https://link3.com", "Memo 3", 2),
	}
	expectedPageAfterRemove := dpage.ReconstructPage("page-1", "Test Page", *creator, "invite-code", expectedLinks, invitedUsers)

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success_by_creator",
			setup: func(m *mocks) {
				links := dpage.Links{
					dpage.ReconstructLink("https://link1.com", "Memo 1", 1),
					dpage.ReconstructLink("https://link2.com", "Memo 2", 2),
					dpage.ReconstructLink("https://link3.com", "Memo 3", 3),
				}
				invitedUsers := duser.Users{invitedUser}
				page := dpage.ReconstructPage("page-1", "Test Page", *creator, "invite-code", links, invitedUsers)
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)
				m.pageRepo.EXPECT().Save(gomock.Any(), expectedPageAfterRemove).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				input: LinkRemoveUsecaseInput{
					PageID: "page-1",
					URL:    "https://link2.com",
				},
			},
			want: want{err: nil},
		},
		{
			name: "success_by_invited_user",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(initialPage, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)
				m.pageRepo.EXPECT().Save(gomock.Any(), expectedPageAfterRemove).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), invitedUser),
				input: LinkRemoveUsecaseInput{
					PageID: "page-1",
					URL:    "https://link2.com",
				},
			},
			want: want{err: nil},
		},
		{
			name: "page_not_found",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "non-existent-id").Return(nil, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				input: LinkRemoveUsecaseInput{
					PageID: "non-existent-id",
					URL:    "https://link2.com",
				},
			},
			want: want{err: ErrPageNotFound},
		},
		{
			name: "user_not_found_in_context",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(initialPage, nil)
			},
			args: args{
				ctx: context.Background(),
				input: LinkRemoveUsecaseInput{
					PageID: "page-1",
					URL:    "https://link2.com",
				},
			},
			want: want{err: ErrUserNotFound},
		},
		{
			name: "unauthorized_user_not_invited",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(initialPage, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), unauthorizedUser),
				input: LinkRemoveUsecaseInput{
					PageID: "page-1",
					URL:    "https://link2.com",
				},
			},
			want: want{err: dpage.ErrNotCreatedByUser},
		},
		{
			name: "link_not_found_on_page",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(initialPage, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return dpage.ErrNotFoundLink("https://non-existent-link.com")
					},
				)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				input: LinkRemoveUsecaseInput{
					PageID: "page-1",
					URL:    "https://non-existent-link.com",
				},
			},
			want: want{err: dpage.ErrNotFoundLink("https://non-existent-link.com")},
		},
		{
			name: "save_error",
			setup: func(m *mocks) {
				links := dpage.Links{
					dpage.ReconstructLink("https://link1.com", "Memo 1", 1),
					dpage.ReconstructLink("https://link2.com", "Memo 2", 2),
					dpage.ReconstructLink("https://link3.com", "Memo 3", 3),
				}
				invitedUsers := duser.Users{invitedUser}
				page := dpage.ReconstructPage("page-1", "Test Page", *creator, "invite-code", links, invitedUsers)
				expectedLinks := dpage.Links{
					dpage.ReconstructLink("https://link1.com", "Memo 1", 1),
					dpage.ReconstructLink("https://link3.com", "Memo 3", 2),
				}
				expectedPageAfterRemove := dpage.ReconstructPage("page-1", "Test Page", *creator, "invite-code", expectedLinks, invitedUsers)
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) },
				)
				m.pageRepo.EXPECT().Save(gomock.Any(), expectedPageAfterRemove).Return(errors.New("save db error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				input: LinkRemoveUsecaseInput{
					PageID: "page-1",
					URL:    "https://link2.com",
				},
			},
			want: want{err: errors.New("save db error")},
		},
		{
			name: "transaction_error",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(initialPage, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(errors.New("txn error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				input: LinkRemoveUsecaseInput{
					PageID: "page-1",
					URL:    "https://link2.com",
				},
			},
			want: want{err: errors.New("txn error")},
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
			tt.setup(m)
			u := NewLinkRemoveUsecase(m.pageRepo, m.txn)
			err := u.LinkRemove(tt.args.ctx, tt.args.input)
			cmperr.Diff(t, tt.want.err, err)
		})
	}
}
