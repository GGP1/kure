package entry

// New creates a new entry.
func New(title, username, password, url, notes, expires string) *Entry {
	return &Entry{
		Title:    title,
		Username: username,
		Password: password,
		URL:      url,
		Notes:    notes,
		Expires:  expires,
	}
}
