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
	mocktransaction "github.com/naka-sei/tsudzuri/usecase/service/mock/mock_transaction"
)

func TestDeleteUsecase_Delete(t *testing.T) {
	type mocks struct {
		pageRepo   *mockpage.MockPageRepository
		txnService *mocktransaction.MockTransactionService
	}
	type args struct {
		ctx    context.Context
		pageID string
	}
	type want struct {
		err error
	}

	user := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)
	other := duser.ReconstructUser("user-id-2", "other-uid", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mocks) {
				m.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
				p1 := dpage.ReconstructPage("page-1", "t", *user, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-1").Return(p1, nil)
				m.pageRepo.EXPECT().DeleteByID(gomock.Any(), "page-1").Return(nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-1",
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "unauthorized",
			setup: func(m *mocks) {
				p2 := dpage.ReconstructPage("page-unauth", "t", *other, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-unauth").Return(p2, nil)
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-unauth",
			},
			want: want{
				err: dpage.ErrNotCreatedByUser,
			},
		},
		{
			name: "delete_error",
			setup: func(m *mocks) {
				m.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					},
				)
				p3 := dpage.ReconstructPage("page-2", "t", *user, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-2").Return(p3, nil)
				m.pageRepo.EXPECT().DeleteByID(gomock.Any(), "page-2").Return(errors.New("delete error"))
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-2",
			},
			want: want{
				err: errors.New("delete error"),
			},
		},
		{
			name: "transaction_error",
			setup: func(m *mocks) {
				p4 := dpage.ReconstructPage("page-3", "t", *user, "invite", dpage.Links{}, duser.Users{})
				m.pageRepo.EXPECT().Get(gomock.Any(), "page-3").Return(p4, nil)
				m.txnService.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Return(errors.New("txn error"))
			},
			args: args{
				ctx:    ctxuser.WithUser(context.Background(), user),
				pageID: "page-3",
			},
			want: want{
				err: errors.New("txn error"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &mocks{
				pageRepo:   mockpage.NewMockPageRepository(ctrl),
				txnService: mocktransaction.NewMockTransactionService(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			u := NewDeleteUsecase(m.pageRepo, m.txnService)
			err := u.Delete(tt.args.ctx, tt.args.pageID)
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
