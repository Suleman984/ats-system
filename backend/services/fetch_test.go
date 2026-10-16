package services

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsPublicIP(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8":         true,
		"127.0.0.1":       false,
		"10.1.2.3":        false,
		"192.168.0.10":    false,
		"169.254.169.254": false,
		"::1":             false,
		"0.0.0.0":         false,
	}
	for ip, want := range cases {
		if got := isPublicIP(net.ParseIP(ip)); got != want {
			t.Errorf("isPublicIP(%s) = %v, want %v", ip, got, want)
		}
	}
}

func TestDownloadBlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("internal"))
	}))
	defer srv.Close()

	if _, _, err := download(context.Background(), srv.URL); err == nil {
		t.Fatal("expected download from a loopback address to be refused")
	}
}

func TestDownloadRejectsNonHTTPSchemes(t *testing.T) {
	for _, u := range []string{"file:///etc/passwd", "gopher://example.com", "ftp://example.com/cv.pdf"} {
		if _, _, err := download(context.Background(), u); err == nil {
			t.Errorf("expected %s to be rejected", u)
		}
	}
}
