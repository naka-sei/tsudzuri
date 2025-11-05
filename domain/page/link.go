package page

import "slices"

type Link struct {
	url      string
	memo     string
	priority int
}

type Links []Link

// AddLink adds a new link to the Links slice.
func (ls *Links) AddLink(url string, memo string) {
	newLink := Link{
		url:      url,
		memo:     memo,
		priority: len(*ls),
	}
	*ls = append(*ls, newLink)
}

// RemoveLink removes a link by its URL.
func (ls *Links) RemoveLink(url string) error {
	deletedIdx, err := ls.getIndexByURL(url)
	if err != nil {
		return err
	}

	// Remove the link
	*ls = slices.Delete(*ls, deletedIdx, deletedIdx+1)

	// Update priorities
	for i := range *ls {
		(*ls)[i].priority = i
	}

	return nil
}

// editLink edits a link's order and memo.
func (ls *Links) editLink(link Link) error {
	slices.SortFunc(*ls, func(a, b Link) int {
		return a.priority - b.priority
	})

	currentIdx, err := ls.getIndexByURL(link.url)
	if err != nil {
		return err
	}

	// Calculate new index
	newIndex := link.priority - 1
	n := len(*ls)
	if newIndex < 0 {
		newIndex = 0
	} else if newIndex >= n {
		newIndex = n - 1
	}

	if newIndex == currentIdx {
		return nil
	}

	// Delete the element from the current position
	*ls = slices.Delete(*ls, currentIdx, currentIdx+1)

	// Insert the updated link at the new position
	*ls = slices.Insert(*ls, newIndex, link)

	// Update priorities
	for i := range *ls {
		(*ls)[i].priority = i + 1
	}

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
