package services

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
)

// fakeClamd accepts one INSTREAM session, checks the framing and replies with the given verdict.
func fakeClamd(t *testing.T, verdict string, gotBody *[]byte) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		cmd := make([]byte, len("zINSTREAM\x00"))
		if _, err := io.ReadFull(conn, cmd); err != nil || string(cmd) != "zINSTREAM\x00" {
			return
		}
		var body bytes.Buffer
		for {
			var size [4]byte
			if _, err := io.ReadFull(conn, size[:]); err != nil {
				return
			}
			n := binary.BigEndian.Uint32(size[:])
			if n == 0 {
				break
			}
			if _, err := io.CopyN(&body, conn, int64(n)); err != nil {
				return
			}
		}
		*gotBody = body.Bytes()
		conn.Write([]byte(verdict + "\x00"))
	}()
	return ln.Addr().String()
}

func TestScanWithClamdClean(t *testing.T) {
	var got []byte
	data := bytes.Repeat([]byte("%PDF"), 40000) // spans several 64 KB chunks
	if err := scanWithClamd(fakeClamd(t, "stream: OK", &got), data); err != nil {
		t.Fatalf("clean file: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("scanner received %d bytes, want %d", len(got), len(data))
	}
}

func TestScanWithClamdInfected(t *testing.T) {
	var got []byte
	err := scanWithClamd(fakeClamd(t, "stream: Eicar-Test-Signature FOUND", &got), []byte("X5O!P%@AP"))
	if !errors.Is(err, ErrInfected) || !strings.Contains(err.Error(), "Eicar-Test-Signature") {
		t.Fatalf("infected file: got %v", err)
	}
}

func TestScanFailsClosedWhenScannerDown(t *testing.T) {
	t.Setenv("CLAMAV_ADDR", "127.0.0.1:1") // nothing listens here
	if err := ScanForMalware([]byte("x")); err == nil || errors.Is(err, ErrInfected) {
		t.Fatalf("unreachable scanner must refuse the upload, got %v", err)
	}
}
