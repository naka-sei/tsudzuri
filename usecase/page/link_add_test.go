package page

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	mockpage "github.com/naka-sei/tsudzuri/domain/page/mock/mock_page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"github.com/naka-sei/tsudzuri/pkg/cmperr"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	mocktxn "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
)

func TestLinkAddUseCase_LinkAdd(t *testing.T) {
	type mocks struct {
		pageRepo *mockpage.MockPageRepository
		txn      *mocktxn.MockTransactionService
	}
	type args struct {
		ctx   context.Context
		input LinkAddUsecaseInput
	}

	creatorUser := duser.ReconstructUser("1", "user1", "anonymous", nil)
	invitedUser := duser.ReconstructUser("2", "user2", "anonymous", nil)
	otherUser := duser.ReconstructUser("3", "user3", "anonymous", nil)

	tests := []struct {
		name    string
		setup   func(m *mocks)
		args    args
		wantErr error
	}{
		{
			name: "success_by_creator",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("1", "Test Page", *creatorUser, "invite-code", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "1").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
						return f(ctx)
					})
				m.pageRepo.EXPECT().Save(gomock.Any(), page).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creatorUser),
				input: LinkAddUsecaseInput{
					PageID: "1",
					URL:    "https://example.com",
					Memo:   "test link",
				},
			},
			wantErr: nil,
		},
		{
			name: "success_by_invited_user",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("1", "Test Page", *creatorUser, "invite-code", dpage.Links{}, duser.Users{invitedUser})
				m.pageRepo.EXPECT().Get(gomock.Any(), "1").Return(page, nil)
				m.txn.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, f func(context.Context) error) error {
						return f(ctx)
					})
				m.pageRepo.EXPECT().Save(gomock.Any(), page).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), invitedUser),
				input: LinkAddUsecaseInput{
					PageID: "1",
					URL:    "https://example.com",
					Memo:   "test link",
				},
			},
			wantErr: nil,
		},
		{
			name: "page_not_found",
			setup: func(m *mocks) {
				m.pageRepo.EXPECT().Get(gomock.Any(), "1").Return(nil, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creatorUser),
				input: LinkAddUsecaseInput{
					PageID: "1",
					URL:    "https://example.com",
					Memo:   "test link",
				},
			},
			wantErr: ErrPageNotFound,
		},
		{
			name: "user_not_found_in_context",
			setup: func(m *mocks) {
				page := &dpage.Page{}
				m.pageRepo.EXPECT().Get(gomock.Any(), "1").Return(page, nil)
			},
			args: args{
				ctx: context.Background(),
				input: LinkAddUsecaseInput{
					PageID: "1",
					URL:    "https://example.com",
					Memo:   "test link",
				},
			},
			wantErr: ErrUserNotFound,
		},
		{
			name: "unauthorized_user",
			setup: func(m *mocks) {
				page := dpage.ReconstructPage("1", "Test Page", *creatorUser, "invite-code", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "1").Return(page, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), otherUser),
				input: LinkAddUsecaseInput{
					PageID: "1",
					URL:    "https://example.com",
					Memo:   "test link",
				},
			},
			wantErr: dpage.ErrNotCreatedByUser,
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
			u := NewLinkAddUsecase(m.pageRepo, m.txn)
			err := u.LinkAdd(tt.args.ctx, tt.args.input)
			cmperr.Diff(t, tt.wantErr, err)
		})
	}
}
