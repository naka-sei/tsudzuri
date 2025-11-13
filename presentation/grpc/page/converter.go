package page

import (
	"math"

	tsudzuriv1 "github.com/naka-sei/tsudzuri/api/tsudzuri/v1"
	dpage "github.com/naka-sei/tsudzuri/domain/page"
)

func toProtoPage(p *dpage.Page) *tsudzuriv1.Page {
	if p == nil {
		return nil
	}

	protoPage := &tsudzuriv1.Page{
		Id:         p.ID(),
		Title:      p.Title(),
		InviteCode: p.InviteCode(),
	}

	if links := p.Links(); len(links) > 0 {
		protoPage.Links = make([]*tsudzuriv1.Link, 0, len(links))
		for _, lnk := range links {
			priority := lnk.Priority()
			var priorityInt32 int32
			if priority < math.MinInt32 || priority > math.MaxInt32 {
				priorityInt32 = 0 // Default to 0 if out of range
			} else {
				priorityInt32 = int32(priority) // #nosec G115 - validated range above
			}
			protoPage.Links = append(protoPage.Links, &tsudzuriv1.Link{
				Url:      lnk.URL(),
				Memo:     lnk.Memo(),
				Priority: priorityInt32,
			})
		}
	}

	return protoPage
}
