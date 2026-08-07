package httphelpers

import (
	"fmt"
	"mime/multipart"
	"net/textproto"
)

func WriteFilePart(w *multipart.Writer, fieldName string, filename string, contentType string, content []byte) error {
	if fieldName == "" || filename == "" || len(content) == 0 {
		return fmt.Errorf("bad file data provided")
	}

	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename)},
		"Content-Type":        {contentType},
	})
	if err != nil {
		return fmt.Errorf("failed to create form part: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		return fmt.Errorf("failed to write part content: %v", err)
	}

	return nil
}
