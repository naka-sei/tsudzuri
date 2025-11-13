package page

import (
	"crypto/rand"
	"fmt"
	"slices"

	duser "github.com/naka-sei/tsudzuri/domain/user"
)

type Page struct {
	id           string
	title        string
	createdBy    duser.User
	inviteCode   string
	links        Links
	invitedUsers duser.Users
}

// NewPage creates a new Page instance.
func NewPage(title string, createdBy *duser.User) (*Page, error) {
	if title == "" {
		return nil, ErrNoTitleProvided
	}

	if createdBy == nil {
		return nil, ErrNoUserProvided
	}

	code, err := inviteCodeGenerator()
	if err != nil {
		return nil, fmt.Errorf("generate invite code: %w", err)
	}

	return &Page{
		title:      title,
		createdBy:  *createdBy,
		inviteCode: code,
		links:      Links{},
	}, nil
}

// ID returns the page's ID.
func (p *Page) ID() string {
	return p.id
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

// CreatedBy returns the creator user of the page.
func (p *Page) CreatedBy() *duser.User {
	return &p.createdBy
}

// InvitedUsers returns the invited users for this page.
func (p *Page) InvitedUsers() duser.Users {
	return p.invitedUsers
}

// Edit edits the page.
func (p *Page) Edit(user *duser.User, title string, links Links) error {
	if err := p.Authorize(user); err != nil {
		return err
	}

	if err := p.links.editLinks(links); err != nil {
		return err
	}

	p.title = title

	return nil
}

// AddLink adds a new link to the page.
func (p *Page) AddLink(user *duser.User, url string, memo string) error {
	if err := p.Authorize(user); err != nil {
		return err
	}

	p.links.addLink(url, memo)
	return nil
}

// RemoveLink removes a link from the page.
func (p *Page) RemoveLink(user *duser.User, url string) error {
	if err := p.Authorize(user); err != nil {
		return err
	}

	if err := p.links.removeLink(url); err != nil {
		return err
	}

	return nil
}

// Authorize authorizes the user to access the page.
func (p *Page) Authorize(user *duser.User) error {
	if user == nil {
		return ErrNoUserProvided
	}

	isUserInvited := slices.ContainsFunc(p.invitedUsers, func(u *duser.User) bool {
		return u.ID() == user.ID()
	})

	if isUserInvited {
		return nil
	}

	if err := p.validateCreatedBy(user); err != nil {
		return err
	}

	return nil
}

// validateCreatedBy validates if the given user is the creator of the page.
func (p *Page) validateCreatedBy(user *duser.User) error {
	if user == nil {
		return ErrNoUserProvided
	}
	if p.createdBy.ID() != user.ID() {
		return ErrNotCreatedByUser
	}
	return nil
}

// ReconstructPage reconstructs a Page instance from existing data.
func ReconstructPage(id string, title string, createdBy duser.User, inviteCode string, links Links, invitedUsers duser.Users) *Page {
	return &Page{
		id:           id,
		title:        title,
		createdBy:    createdBy,
		inviteCode:   inviteCode,
		links:        links,
		invitedUsers: invitedUsers,
	}
}

var inviteCodeGenerator = defaultInviteCodeGenerator

const inviteCodeLength = 8
const inviteCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func defaultInviteCodeGenerator() (string, error) {
	buf := make([]byte, inviteCodeLength)
	max := byte(len(inviteCodeAlphabet))
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = inviteCodeAlphabet[b%max]
	}
	return string(buf), nil
}
