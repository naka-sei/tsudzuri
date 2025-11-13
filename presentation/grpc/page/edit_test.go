package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mockedit "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_edit"
)

func TestEditService_Edit(t *testing.T) {
	type args struct {
		ctx context.Context
		req *tsudzuriv1.EditPageRequest
	}
	type want struct {
		res *emptypb.Empty
		err error
	}

	user := duser.ReconstructUser("user-id", "uid-1", "anonymous", nil)

	tests := []struct {
		name  string
		setup func(m *mockedit.MockEditUsecase)
		args  args
		want  want
	}{
		{
			name: "success",
			setup: func(m *mockedit.MockEditUsecase) {
				expected := dpage.Links{
					dpage.ReconstructLink("https://example.com", "memo", 1),
				}
				m.EXPECT().Edit(gomock.Any(), "page-1", "new-title", expected).Return(nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.EditPageRequest{
					PageId: "page-1",
					Title:  "new-title",
					Links: []*tsudzuriv1.LinkInput{{
						Url:      "https://example.com",
						Memo:     "memo",
						Priority: 1,
					}},
				},
			},
			want: want{
				res: &emptypb.Empty{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mockedit.MockEditUsecase) {
				expected := dpage.Links{
					dpage.ReconstructLink("https://example.com", "memo", 1),
				}
				m.EXPECT().Edit(gomock.Any(), "page-1", "new-title", expected).Return(errors.New("edit error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), user),
				req: &tsudzuriv1.EditPageRequest{
					PageId: "page-1",
					Title:  "new-title",
					Links: []*tsudzuriv1.LinkInput{{
						Url:      "https://example.com",
						Memo:     "memo",
						Priority: 1,
					}},
				},
			},
			want: want{
				res: nil,
				err: errors.New("edit error"),
			},
		},
		{
			name:  "user_not_found",
			setup: nil,
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.EditPageRequest{PageId: "page-1"},
			},
			want: want{
				res: nil,
				err: duser.ErrUserNotFound,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			usecase := mockedit.NewMockEditUsecase(ctrl)
			if tt.setup != nil {
				tt.setup(usecase)
			}

			svc := NewEditService(usecase)
			got, err := svc.Edit(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got, protocmp.Transform()); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
