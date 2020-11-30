package payload

import (
	"io"

	"encoding/base64"
)

// Encoding is a supported Document encoding type
type Encoding string

const (
	// NoEncoding is the default document data encoding, which should be json
	// serializable.
	NoEncoding Encoding = ""
	// Base64Encoding is the default encoding for binary documents, so it can be
	// json compatible
	Base64Encoding Encoding = "base64"
)

// Documents contains a set of Documents by name
type Documents map[string]Document

// Document is a payload item containing the payload data, content type, and
// optional encoding.
type Document struct {
	ContentType string `json:"content_type"`
	Data        string `json:"data"`
	readIndex   int64
	Encoding    Encoding `json:"encoding,omitempty"`
}

// NewDocument creates a new document of string data with content type, and
// encoding.
func NewDocument(data string, ct string, enc Encoding) Document {
	return Document{
		ContentType: ct,
		Data:        data,
		Encoding:    enc,
	}
}

// InitDocument starts a blank document with the content type and desired
// encoding.
func InitDocument(ct string, enc Encoding) Document {
	return Document{
		ContentType: ct,
		Data:        "",
		Encoding:    enc,
	}
}

// Reader returns an io.Reader for the document data with the appropriate
// encoding applied
func (d *Document) Reader() (r io.Reader) {
	r = d

	if d.Encoding == Base64Encoding {
		r = base64.NewDecoder(base64.StdEncoding, d)
	}

	return
}

// Read returns a slice of the Document data verbatim. It does not apply the
// appropriate; use the Reader().Read() method for this purpose.
func (d *Document) Read(p []byte) (n int, err error) {
	if d.readIndex >= int64(len([]byte(d.Data))) {
		err = io.EOF
		return
	}

	n = copy(p, []byte(d.Data)[d.readIndex:])
	d.readIndex += int64(n)
	return
}

// WriteCloser returns a stream writer that applies the appropriate encoding.
// When finished writing, the caller must Close the returned encoder to flush
// any partially written data.
func (d *Document) WriteCloser() (w io.WriteCloser) {
	w = d

	if d.Encoding == Base64Encoding {
		w = base64.NewEncoder(base64.StdEncoding, d)
	}

	return w
}

// Write appends the provided bytes to the Document data verbatim. It does not
// apply the appropriate; use the WriteCloser method for this purpose.
func (d *Document) Write(data []byte) (n int, err error) {
	d.Data += string(data)
	return len(data), nil
}

// Close does nothing. It exists to adhere to the WriteCloser interface.
func (d Document) Close() error {
	return nil
}
