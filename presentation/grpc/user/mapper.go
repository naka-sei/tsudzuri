package user

import (
	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	duser "github.com/naka-sei/tsudzuri/domain/user"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func toProtoUser(u *duser.User) *tsudzuriv1.User {
	if u == nil {
		return nil
	}
	proto := &tsudzuriv1.User{
		Id:            u.ID(),
		Uid:           u.UID(),
		Provider:      string(u.Provider()),
		JoinedPageIds: u.JoinedPageIDs(),
	}
	if email := u.Email(); email != nil {
		proto.Email = wrapperspb.String(*email)
	}
	if proto.JoinedPageIds == nil {
		proto.JoinedPageIds = []string{}
	}
	return proto
}
