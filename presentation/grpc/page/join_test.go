package page

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	mockjoinusecase "github.com/naka-sei/tsudzuri/usecase/page/mock/mock_join"
)

func TestJoinService_Join(t *testing.T) {
	type fields struct {
		joinUsecase *mockjoinusecase.MockJoinUsecase
	}
	type args struct {
		ctx context.Context
		req *tsudzuriv1.JoinPageRequest
	}
	type want struct {
		resp *emptypb.Empty
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
				f.joinUsecase.EXPECT().Join(gomock.Any(), "page-id", "INVITE01").Return(nil)
			},
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.JoinPageRequest{PageId: "page-id", InviteCode: "INVITE01"},
			},
			want: want{resp: &emptypb.Empty{}, err: nil},
		},
		{
			name: "usecase_error",
			setup: func(f *fields) {
				f.joinUsecase.EXPECT().Join(gomock.Any(), "page-id", "INVITE01").Return(errors.New("join error"))
			},
			args: args{
				ctx: context.Background(),
				req: &tsudzuriv1.JoinPageRequest{PageId: "page-id", InviteCode: "INVITE01"},
			},
			want: want{resp: nil, err: errors.New("join error")},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := &fields{
				joinUsecase: mockjoinusecase.NewMockJoinUsecase(ctrl),
			}
			if tt.setup != nil {
				tt.setup(f)
			}

			svc := NewJoinService(f.joinUsecase)
			resp, err := svc.Join(tt.args.ctx, tt.args.req)

			if tt.want.resp == nil {
				if resp != nil {
					t.Fatalf("expected nil response, got %#v", resp)
				}
			} else {
				if resp == nil {
					t.Fatalf("expected response, got nil")
				}
			}

			if (err == nil) != (tt.want.err == nil) {
				t.Fatalf("error mismatch: want %v, got %v", tt.want.err, err)
			}
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Fatalf("error mismatch: want %v, got %v", tt.want.err, err)
			}
		})
	}
}
