package file

// New creates a new file.
func New(name, filename string, content []byte, createdAt int64) *File {
	return &File{
		Name:      name,
		Content:   content,
		Filename:  filename,
		CreatedAt: createdAt,
	}
}
