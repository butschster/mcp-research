package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The embedded frontend was served raw. That was tolerable while the largest
// asset was a few hundred kilobytes; the API reference brought a 2.3 MB chunk
// that gzips to 670 KB, on a page whose whole audience is remote.
func TestStatic_CompressesWhatItCan(t *testing.T) {
	h := staticHandler()

	get := func(path, acceptEncoding string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		if acceptEncoding != "" {
			r.Header.Set("Accept-Encoding", acceptEncoding)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}

	// The test fixture's static directory holds only a stub index.html, so this
	// asserts the mechanism rather than a ratio.
	plain := get("/index.html", "")
	if plain.Code != http.StatusOK {
		t.Fatalf("index.html: status %d", plain.Code)
	}
	if enc := plain.Header().Get("Content-Encoding"); enc != "" {
		t.Errorf("a client that did not ask for gzip got Content-Encoding %q", enc)
	}
	if plain.Header().Get("Vary") != "Accept-Encoding" {
		t.Error("no Vary: Accept-Encoding — a shared cache would serve one encoding to everybody")
	}

	zipped := get("/index.html", "gzip, deflate, br")
	if zipped.Code != http.StatusOK {
		t.Fatalf("index.html gzip: status %d", zipped.Code)
	}

	// A tiny stub may not shrink, and then it is correctly served raw. Whatever
	// the answer, the encoding header and the body have to agree — claiming
	// gzip over plain bytes is the failure that breaks every client at once.
	if zipped.Header().Get("Content-Encoding") == "gzip" {
		zr, err := gzip.NewReader(strings.NewReader(zipped.Body.String()))
		if err != nil {
			t.Fatalf("body announced as gzip does not decompress: %v", err)
		}
		out, err := io.ReadAll(zr)
		if err != nil {
			t.Fatalf("reading the gzip body: %v", err)
		}
		if string(out) != plain.Body.String() {
			t.Error("the compressed body decompresses to something other than the raw one")
		}
		if zipped.Body.Len() >= plain.Body.Len() {
			t.Error("a body was compressed into something larger and still announced as gzip")
		}
	}

	// Content-Type must survive compression: a JavaScript chunk served as
	// text/plain is a page that silently does not run.
	if ct := zipped.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type %q", ct)
	}
}

// A second request must serve the cached bytes, not a fresh compression, and
// must be byte-identical — a cache that returns something different under
// concurrency is worse than no cache.
func TestStatic_CompressionIsStable(t *testing.T) {
	h := staticHandler()
	body := func() string {
		r := httptest.NewRequest("GET", "/index.html", nil)
		r.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Body.String()
	}
	first := body()
	for i := 0; i < 5; i++ {
		if body() != first {
			t.Fatal("the same asset came back differently on a repeat request")
		}
	}
}

// Compressing an image or a font costs time to make the file slightly bigger.
func TestStatic_LeavesAlreadyCompressedFormatsAlone(t *testing.T) {
	for _, ext := range []string{".png", ".woff2", ".ico", ".webp", ".jpg", ".zip"} {
		if compressible(ext) {
			t.Errorf("%s is treated as compressible", ext)
		}
	}
	for _, ext := range []string{".js", ".css", ".html", ".json", ".svg"} {
		if !compressible(ext) {
			t.Errorf("%s is not treated as compressible", ext)
		}
	}
}
