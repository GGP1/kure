package entry

// New creates a new entry.
func New(name, username, password, url, notes, expires string) *Entry {
	return &Entry{
		Name:     name,
		Username: username,
		Password: password,
		URL:      url,
		Notes:    notes,
		Expires:  expires,
	}
}
