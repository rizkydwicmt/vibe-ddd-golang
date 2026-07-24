package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"reflect"
	"testing"

	typeBuf "vibe-ddd-golang/internal/common/type/file"
)

func TestCreateJSONBody(t *testing.T) {
	input := map[string]interface{}{"foo": "bar", "baz": float64(123)}
	reader, err := createJSONBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(reader).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !reflect.DeepEqual(input, out) {
		t.Errorf("expected %v, got %v", input, out)
	}
}

func TestCreateFormURLEncodedBody(t *testing.T) {
	input := map[string]string{"foo": "bar", "baz": "qux"}
	reader, err := createFormURLEncodedBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	result := buf.String()
	if !(result == "baz=qux&foo=bar" || result == "foo=bar&baz=qux") {
		t.Errorf("unexpected form result: %s", result)
	}
}

func TestCreateFormURLEncodedBody_InvalidType(t *testing.T) {
	_, err := createFormURLEncodedBody([]string{"foo", "bar"})
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}

func TestCreateMultipartBody_StringAndBytes(t *testing.T) {
	input := map[string]interface{}{
		"field1": "value1",
		"file1":  []byte("filecontent"),
	}
	reader, contentType, err := createMultipartBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.HasPrefix([]byte(contentType), []byte("multipart/form-data")) {
		t.Errorf("unexpected content type: %s", contentType)
	}
	// Parse multipart
	mr := multipart.NewReader(reader, contentType[len("multipart/form-data; boundary="):])
	fields := map[string]bool{"field1": false, "file1": false}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("multipart read error: %v", err)
		}
		if part.FormName() == "field1" {
			val, _ := io.ReadAll(part)
			if string(val) != "value1" {
				t.Errorf("expected value1, got %s", string(val))
			}
			fields["field1"] = true
		}
		if part.FormName() == "file1" {
			val, _ := io.ReadAll(part)
			if string(val) != "filecontent" {
				t.Errorf("expected filecontent, got %s", string(val))
			}
			fields["file1"] = true
		}
	}
	for k, v := range fields {
		if !v {
			t.Errorf("missing field: %s", k)
		}
	}
}

func TestCreateMultipartBody_BufferedFile(t *testing.T) {
	file := typeBuf.BufferedFile{
		OriginalName: "playground.txt",
		MimeType:     "text/plain",
		Encoding:     "utf-8",
		Size:         len([]byte("hello")),
		Buffer:       []byte("hello"),
	}
	input := map[string]interface{}{
		"file": file,
	}
	reader, contentType, err := createMultipartBody(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.HasPrefix([]byte(contentType), []byte("multipart/form-data")) {
		t.Errorf("unexpected content type: %s", contentType)
	}
	// Parse multipart
	mr := multipart.NewReader(reader, contentType[len("multipart/form-data; boundary="):])
	found := false
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("multipart read error: %v", err)
		}
		if part.FormName() == "file" {
			val, _ := io.ReadAll(part)
			if string(val) != "hello" {
				t.Errorf("expected hello, got %s", string(val))
			}
			found = true
		}
	}
	if !found {
		t.Error("file part not found in multipart body")
	}
}

func TestCreateMultipartBody_InvalidType(t *testing.T) {
	_, _, err := createMultipartBody([]string{"foo", "bar"})
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}
