package handlers

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"uuid"

	"github.com/andrey1af/shop-api/api-service/internal/models"
)

func TestReadImageRejectsEmptyBody(t *testing.T) {
	request := httptest.NewRequest("POST", "/", nil)
	recorder := httptest.NewRecorder()

	if _, ok := readImage(recorder, request); ok {
		t.Fatal("readImage() accepted an empty body")
	}
	if recorder.Code != 400 {
		t.Fatalf("readImage() status = %d, want 400", recorder.Code)
	}
}

func TestWriteImageReturnsDownload(t *testing.T) {
	image := models.Image{ID: uuid.New(), Data: []byte{0x01, 0x02, 0x03}}
	recorder := httptest.NewRecorder()

	writeImage(recorder, image)
	response := recorder.Result()
	t.Cleanup(func() {
		if err := response.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	})

	if response.Header.Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("Content-Type = %q", response.Header.Get("Content-Type"))
	}
	if !strings.HasPrefix(response.Header.Get("Content-Disposition"), "attachment;") {
		t.Fatalf("Content-Disposition = %q", response.Header.Get("Content-Disposition"))
	}
	if !bytes.Equal(recorder.Body.Bytes(), image.Data) {
		t.Fatalf("body = %v, want %v", recorder.Body.Bytes(), image.Data)
	}
}
