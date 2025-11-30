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
	mockget "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_get"
)

func TestGetService_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		req *tsudzuriv1.GetPageRequest
	}
	type want struct {
		res *tsudzuriv1.Page
		err error
	}

	creator := duser.ReconstructUser("creator-id", "uid-1", "anonymous", nil)
	invited := duser.ReconstructUser("invited-id", "uid-2", "anonymous", nil)

	pageWithoutLinks := dpage.ReconstructPage("page-1", "title-1", *creator, "invite-code", nil, nil)
	pageWithLinks := dpage.ReconstructPage("page-2", "title-2", *creator, "invite-code", dpage.Links{
		dpage.ReconstructLink("https://example.com", "memo", 1),
	}, duser.Users{invited})

	tests := []struct {
		name  string
		setup func(m *mockget.MockGetUsecase)
		args  args
		want  want
	}{
		{
			name: "success_without_links",
			setup: func(m *mockget.MockGetUsecase) {
				m.EXPECT().Get(gomock.Any(), "page-1").Return(pageWithoutLinks, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: &tsudzuriv1.GetPageRequest{PageId: "page-1"},
			},
			want: want{
				res: &tsudzuriv1.Page{
					Id:         "page-1",
					Title:      "title-1",
					InviteCode: "invite-code",
				},
				err: nil,
			},
		},
		{
			name: "success_with_links",
			setup: func(m *mockget.MockGetUsecase) {
				m.EXPECT().Get(gomock.Any(), "page-2").Return(pageWithLinks, nil)
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), invited),
				req: &tsudzuriv1.GetPageRequest{PageId: "page-2"},
			},
			want: want{
				res: &tsudzuriv1.Page{
					Id:         "page-2",
					Title:      "title-2",
					InviteCode: "",
					Links: []*tsudzuriv1.Link{{
						Url:      "https://example.com",
						Memo:     "memo",
						Priority: 1,
					}},
				},
				err: nil,
			},
		},
		{
			name: "usecase_error",
			setup: func(m *mockget.MockGetUsecase) {
				m.EXPECT().Get(gomock.Any(), "page-1").Return(nil, errors.New("get error"))
			},
			args: args{
				ctx: ctxuser.WithUser(context.Background(), creator),
				req: &tsudzuriv1.GetPageRequest{PageId: "page-1"},
			},
			want: want{
				res: nil,
				err: errors.New("get error"),
			},
		},
		{
			name:  "user_not_found",
			setup: nil,
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.GetPageRequest{PageId: "page-1"},
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

			usecase := mockget.NewMockGetUsecase(ctrl)
			if tt.setup != nil {
				tt.setup(usecase)
			}

			svc := NewGetService(usecase)
			got, err := svc.Get(tt.args.ctx, tt.args.req)
			if diff := cmp.Diff(tt.want.res, got, protocmp.Transform()); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
			testutil.EqualErr(t, tt.want.err, err)
		})
	}
}
