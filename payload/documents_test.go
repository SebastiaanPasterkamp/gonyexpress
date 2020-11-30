package payload_test

import (
	"github.com/SebastiaanPasterkamp/gonyexpress/payload"

	"testing"
)

var newDocumentCases = []struct {
	Name        string
	Data        string
	ContentType string
	Encoding    payload.Encoding

	ExpectedData string
	ExpectedCT   string
	ExpectedEnc  payload.Encoding
}{
	{"Plain", "Foo bar!", "text/plain", payload.NoEncoding,
		"Foo bar!", "text/plain", payload.NoEncoding},
	{"Pre encoded", "PHhtbD5Gb28gYmFyJmFtcDs8L3htbD4=", "text/xml", payload.Base64Encoding,
		"PHhtbD5Gb28gYmFyJmFtcDs8L3htbD4=", "text/xml", payload.Base64Encoding},
}

func TestNewDocumentst(t *testing.T) {
	for _, tc := range newDocumentCases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			d := payload.NewDocument(tc.Data, tc.ContentType, tc.Encoding)

			if d.ContentType != tc.ExpectedCT {
				t.Errorf("ContentType = %+v; expected %+v",
					d.Encoding, tc.ExpectedCT)
			}

			if d.Encoding != tc.ExpectedEnc {
				t.Errorf("Encoding = %+v; expected %+v",
					d.Encoding, tc.ExpectedEnc)
			}

			if d.Data != tc.ExpectedData {
				t.Errorf("Data = '%s'; expected '%s'",
					d.Data, tc.ExpectedData)
			}
		})
	}
}

func TestDocumentWriting(t *testing.T) {
	var writeDocumentCases = []struct {
		Name        string
		Data        string
		ContentType string
		Encoding    payload.Encoding

		ExpectedData string
		ExpectedCT   string
		ExpectedEnc  payload.Encoding
	}{
		{"Plain", "Foo bar!", "text/plain", payload.NoEncoding,
			"Foo bar!", "text/plain", payload.NoEncoding},
		{"Auto encode", "<xml>Foo bar&amp;</xml>", "text/xml", payload.Base64Encoding,
			"PHhtbD5Gb28gYmFyJmFtcDs8L3htbD4=", "text/xml", payload.Base64Encoding},
	}

	for _, tc := range writeDocumentCases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			d := payload.InitDocument(
				tc.ContentType,
				tc.Encoding,
			)

			w := d.WriteCloser()
			n, err := w.Write([]byte(tc.Data))
			if err != nil {
				t.Fatalf("Unexpected error writing %+v: %+v", tc.Data, err)
			}
			if n != len(tc.Data) {
				t.Errorf("Unexpected number of bytes written; Wrote %d, expected %d.",
					n, len(tc.Data))
			}
			w.Close()

			if d.ContentType != tc.ExpectedCT {
				t.Errorf("ContentType = %+v; expected %+v",
					d.Encoding, tc.ExpectedCT)
			}

			if d.Encoding != tc.ExpectedEnc {
				t.Errorf("Encoding = %+v; expected %+v",
					d.Encoding, tc.ExpectedEnc)
			}

			if d.Data != tc.ExpectedData {
				t.Errorf("Data = '%s'; expected '%s'",
					d.Data, tc.ExpectedData)
			}
		})
	}
}

func TestDocumentReading(t *testing.T) {
	var readDocumentCases = []struct {
		Name        string
		Data        string
		ContentType string
		Encoding    payload.Encoding

		ExpectedData string
		ExpectedCT   string
		ExpectedEnc  payload.Encoding
	}{
		{"Plain", "Foo bar!", "text/plain", payload.NoEncoding,
			"Foo bar!", "text/plain", payload.NoEncoding},
		{"Auto decode", "PHhtbD5Gb28gYmFyJmFtcDs8L3htbD4=", "text/xml", payload.Base64Encoding,
			"<xml>Foo bar&amp;</xml>", "text/xml", payload.Base64Encoding},
	}

	for _, tc := range readDocumentCases {
		tc := tc // capture range variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			d := payload.NewDocument(tc.Data, tc.ContentType, tc.Encoding)

			buf := make([]byte, 32)
			n, err := d.Reader().Read(buf)
			if err != nil {
				t.Fatalf("Unexpected error writing %+v: %+v", tc.Data, err)
			}
			if n != len(tc.ExpectedData) {
				t.Errorf("Unexpected number of bytes read; Got %d, expected %d.",
					n, len(tc.ExpectedData))
			}

			if d.ContentType != tc.ExpectedCT {
				t.Errorf("ContentType = %+v; expected %+v",
					d.Encoding, tc.ExpectedCT)
			}

			if d.Encoding != tc.ExpectedEnc {
				t.Errorf("Encoding = %+v; expected %+v",
					d.Encoding, tc.ExpectedEnc)
			}

			if string(buf[:n]) != tc.ExpectedData {
				t.Errorf("Data = '%s'; expected '%s'",
					string(buf[:n]), tc.ExpectedData)
			}

			n, err = d.Reader().Read(buf)
			if err == nil {
				t.Fatalf("Expected EOF, inread read %d bytes: '%s'", n, buf)
			}
		})
	}
}
