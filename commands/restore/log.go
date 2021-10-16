package restore

import (
	"encoding/binary"
	"io"
	"os"

	cmdutil "github.com/GGP1/kure/commands"
)

// log represents a log file.
type log struct {
	bucketName []byte
	file       *os.File
	closed     bool
}

// newLog creates a new write-ahead log.
func newLog(bucketName []byte) (*log, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	l := &log{
		bucketName: bucketName,
		file:       f,
		closed:     false,
	}
	return l, nil
}

// BucketName returns the name of the bucket that the log is persisting.
func (l *log) BucketName() []byte {
	return l.bucketName
}

// Close closes and erases the log file.
func (l *log) Close() error {
	if l.closed {
		return nil
	}
	if err := l.Sync(); err != nil {
		return err
	}
	if err := l.file.Close(); err != nil {
		return err
	}
	l.closed = true
	return cmdutil.Erase(l.file.Name())
}

// Read reads the log and returns a slice of records.
func (l *log) Read() ([][]byte, error) {
	if l.closed {
		return nil, os.ErrClosed
	}

	stat, err := l.file.Stat()
	if err != nil {
		return nil, err
	}

	records := make([][]byte, 0)
	numSize := int64(binary.MaxVarintLen64)
	num := make([]byte, numSize) // Reuse

	for i := int64(0); i < stat.Size(); {
		if _, err := l.file.ReadAt(num, i); err != nil {
			return nil, err
		}

		recordSize, _ := binary.Uvarint(num)
		record := make([]byte, recordSize)

		if _, err := l.file.ReadAt(record, i+numSize); err != nil && err != io.EOF {
			return nil, err
		}

		records = append(records, record)
		i += int64(recordSize) + numSize
	}

	return records, nil
}

// Sync commits the current contents of the file to stable storage.
// Typically, this means flushing the file system's in-memory copy of recently written data to disk.
func (l *log) Sync() error {
	return l.file.Sync()
}

// Write writes data to the log in the form: dataSize + data.
func (l *log) Write(data []byte) error {
	if l.closed {
		return os.ErrClosed
	}

	dataSize := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(dataSize, uint64(len(data)))
	data = append(dataSize, data...)

	if _, err := l.file.Write(data); err != nil {
		return err
	}

	return nil
}
