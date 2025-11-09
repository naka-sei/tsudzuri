package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	"github.com/naka-sei/tsudzuri/presentation/http/response"
	mockedit "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_edit"
)

func TestEditService_Edit(t *testing.T) {
	type mocks struct {
		editUsecase *mockedit.MockEditUsecase
	}
	type args struct {
		ctx context.Context
		req EditRequest
	}
	type want struct {
		res response.EmptyResponse
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
				links := dpage.Links{dpage.ReconstructLink("url1", "memo1", 1)}
				m.editUsecase.EXPECT().Edit(gomock.Any(), "page-1", "new title", links).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: EditRequest{
					PageID: "page-1",
					Title:  "new title",
					Links: []LinkRequest{
						{URL: "url1", Memo: "memo1", Priority: 1},
					},
				},
			},
			want: want{
				res: response.EmptyResponse{},
				err: nil,
			},
		},
		{
			name: "success_no_links",
			setup: func(m *mocks) {
				m.editUsecase.EXPECT().Edit(gomock.Any(), "page-1", "new title", gomock.Nil()).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: EditRequest{
					PageID: "page-1",
					Title:  "new title",
					Links:  []LinkRequest{},
				},
			},
			want: want{
				res: response.EmptyResponse{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocks) {
				m.editUsecase.EXPECT().Edit(gomock.Any(), "page-1", "new title", gomock.Nil()).Return(errors.New("edit error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: EditRequest{
					PageID: "page-1",
					Title:  "new title",
					Links:  []LinkRequest{},
				},
			},
			want: want{
				res: response.EmptyResponse{},
				err: errors.New("edit error"),
			},
		},
		{
			name: "user_not_found",
			setup: func(m *mocks) {
				m.editUsecase.EXPECT().Edit(gomock.Any(), "page-1", "new title", gomock.Nil()).Return(duser.ErrUserNotFound)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: EditRequest{
					PageID: "page-1",
					Title:  "new title",
					Links:  []LinkRequest{},
				},
			},
			want: want{
				res: response.EmptyResponse{},
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
				editUsecase: mockedit.NewMockEditUsecase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(m)
			}
			s := NewEditService(m.editUsecase)
			got, err := s.Edit(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
