package file

// New creates a new file.
func New(name, typ string, content []byte, createdAt int64) *File {
	return &File{
		Name:      name,
		Content:   content,
		Type:      typ,
		CreatedAt: createdAt,
	}
}
