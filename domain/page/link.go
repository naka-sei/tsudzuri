package page

import "slices"

type Link struct {
	url      string
	memo     string
	priority int
}

// URL returns the link URL.
func (l Link) URL() string { return l.url }

// Memo returns the link memo.
func (l Link) Memo() string { return l.memo }

// Priority returns the link priority.
func (l Link) Priority() int { return l.priority }

type Links []Link

// addLink adds a new link to the Links slice.
func (ls *Links) addLink(url string, memo string) {
	newLink := Link{
		url:      url,
		memo:     memo,
		priority: len(*ls),
	}
	*ls = append(*ls, newLink)
}

// removeLink removes a link by its URL.
func (ls *Links) removeLink(url string) error {
	deletedIdx, err := ls.getIndexByURL(url)
	if err != nil {
		return err
	}

	// Remove the link
	*ls = slices.Delete(*ls, deletedIdx, deletedIdx+1)

	// Update priorities
	for i := range *ls {
		(*ls)[i].priority = i + 1
	}

	return nil
}

// editLink edits a link's order and memo.
func (ls *Links) editLinks(links Links) error {
	if len(links) != len(*ls) {
		return ErrInvalidLinksLength
	}

	slices.SortFunc(links, func(a, b Link) int {
		return a.priority - b.priority
	})

	for i, link := range links {
		_, err := ls.getIndexByURL(link.url)
		if err != nil {
			return err
		}
		links[i].priority = i + 1
	}

	*ls = links

	return nil
}

// getIndexByURL returns the index of the link with the given URL.
func (ls Links) getIndexByURL(url string) (int, error) {
	idx := slices.IndexFunc(ls, func(l Link) bool {
		return l.url == url
	})
	if idx == -1 {
		return -1, ErrNotFoundLink(url)
	}
	return idx, nil
}

// ReconstructLink reconstructs a Link from its components.
func ReconstructLink(url string, memo string, priority int) Link {
	return Link{
		url:      url,
		memo:     memo,
		priority: priority,
	}
}
