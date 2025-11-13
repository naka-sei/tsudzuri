package page

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	dpage "github.com/naka-sei/tsudzuri/domain/page"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	ctxuser "github.com/naka-sei/tsudzuri/pkg/ctx/user"
	"github.com/naka-sei/tsudzuri/pkg/testutil"
	mocklist "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_list"
)

func TestListService_List(t *testing.T) {
	type args struct {
		ctx context.Context
		req *tsudzuriv1.ListPagesRequest
	}
	type want struct {
		res *tsudzuriv1.ListPagesResponse
		err error
	}

	creator := duser.ReconstructUser("creator-id", "uid-1", "anonymous", nil)

	page1 := dpage.ReconstructPage("page-1", "title-1", *creator, "code-1", nil, nil)
	page2 := dpage.ReconstructPage("page-2", "title-2", *creator, "code-2", dpage.Links{
		dpage.ReconstructLink("https://example.com", "memo", 1),
	}, nil)

	tests := []struct {
		name  string
		setup func(m *mocklist.MockListUsecase)
		args  args
		want  want
	}{
		{
			name: "success_with_pages",
			setup: func(m *mocklist.MockListUsecase) {
				m.EXPECT().List(gomock.Any()).Return([]*dpage.Page{page1, page2}, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: &tsudzuriv1.ListPagesRequest{},
			},
			want: want{
				res: &tsudzuriv1.ListPagesResponse{
					Pages: []*tsudzuriv1.Page{
						{
							Id:         "page-1",
							Title:      "title-1",
							InviteCode: "code-1",
						},
						{
							Id:         "page-2",
							Title:      "title-2",
							InviteCode: "code-2",
							Links: []*tsudzuriv1.Link{{
								Url:      "https://example.com",
								Memo:     "memo",
								Priority: 1,
							}},
						},
					},
				},
				err: nil,
			},
		},
		{
			name: "success_no_pages",
			setup: func(m *mocklist.MockListUsecase) {
				m.EXPECT().List(gomock.Any()).Return([]*dpage.Page{}, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: &tsudzuriv1.ListPagesRequest{},
			},
			want: want{
				res: &tsudzuriv1.ListPagesResponse{},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mocklist.MockListUsecase) {
				m.EXPECT().List(gomock.Any()).Return(nil, errors.New("list error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: &tsudzuriv1.ListPagesRequest{},
			},
			want: want{
				res: nil,
				err: errors.New("list error"),
			},
		},
		{
			name:  "user_not_found",
			setup: nil,
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.ListPagesRequest{},
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

			usecase := mocklist.NewMockListUsecase(ctrl)
			if tt.setup != nil {
				tt.setup(usecase)
			}

			svc := NewListService(usecase)
			got, err := svc.List(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got, protocmp.Transform()); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
