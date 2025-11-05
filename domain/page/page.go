package page

import (
	di "github.com/naka-sei/tsudzuri/domain/user"
)

type Page struct {
	title      string
	createdBy  di.User
	inviteCode string
	links      Links
}

// NewPage creates a new Page instance.
func NewPage(title string, createdBy di.User, inviteCode string) *Page {
	return &Page{
		title:      title,
		createdBy:  createdBy,
		inviteCode: inviteCode,
		links:      Links{},
	}
}

// Title returns the page's title.
func (p *Page) Title() string {
	return p.title
}

// inviteCode returns the page's invite code.
func (p *Page) InviteCode() string {
	return p.inviteCode
}

// Links returns the page's links.
func (p *Page) Links() Links {
	return p.links
}

// Edit edits the page.
func (p *Page) Edit(title string, links Links) error {
	if len(links) != len(p.links) {
		return ErrInvalidLinksLength
	}

	p.title = title

	for _, link := range links {
		if err := p.links.editLink(link); err != nil {
			return err
		}
	}

	return nil
}
