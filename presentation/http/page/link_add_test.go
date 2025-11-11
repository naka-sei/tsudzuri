package page

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"

	upage "github.com/naka-sei/tsudzuri/usecase/page"
	mocklinkadd "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_link_add"
)

func TestLinkAddService_LinkAdd(t *testing.T) {
	type mocks struct {
		linkAddUsecase *mocklinkadd.MockLinkAddUseCase
	}
	type args struct {
		ctx context.Context
		req LinkAddRequest
	}
	type want struct {
		err error
	}

	user := duser.ReconstructUser("user-id-1", "uid-1", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(m *mocks)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mocks) {
				input := upage.LinkAddUsecaseInput{PageID: "page-1", URL: "https://example.com", Memo: "test memo"}
				m.linkAddUsecase.EXPECT().LinkAdd(gomock.Any(), input).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: LinkAddRequest{
					PageID: "page-1",
					URL:    "https://example.com",
					Memo:   "test memo",
				},
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				input := upage.LinkAddUsecaseInput{PageID: "page-1", URL: "https://example.com", Memo: "test memo"}
				m.linkAddUsecase.EXPECT().LinkAdd(gomock.Any(), input).Return(errors.New("link add error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: LinkAddRequest{
					PageID: "page-1",
					URL:    "https://example.com",
					Memo:   "test memo",
				},
			},
			want: want{
				err: errors.New("link add error"),
			},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				input := upage.LinkAddUsecaseInput{PageID: "page-1", URL: "https://example.com", Memo: "test memo"}
				m.linkAddUsecase.EXPECT().LinkAdd(gomock.Any(), input).Return(duser.ErrUserNotFound)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: LinkAddRequest{
					PageID: "page-1",
					URL:    "https://example.com",
					Memo:   "test memo",
				},
			},
			want: want{
				err: duser.ErrUserNotFound,
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
				linkAddUsecase: mocklinkadd.NewMockLinkAddUseCase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewLinkAddService(m.linkAddUsecase)
			err := s.LinkAdd(tt.args.ctx, tt.args.req)
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
