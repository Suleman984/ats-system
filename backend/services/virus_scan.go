package services

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Uploads are scanned by ClamAV (clamd) over its INSTREAM protocol when CLAMAV_ADDR is set,
// e.g. "localhost:3310". While enabled, a file that can't be scanned is refused, not let through.

// ErrInfected means clamd recognised malware in the file.
var ErrInfected = errors.New("file is infected")

const (
	clamTimeout   = 30 * time.Second
	clamChunkSize = 64 << 10
)

var warnScanDisabled sync.Once

func ScanForMalware(data []byte) error {
	addr := os.Getenv("CLAMAV_ADDR")
	if addr == "" {
		warnScanDisabled.Do(func() { log.Print("CLAMAV_ADDR not set: uploaded files are not virus-scanned") })
		return nil
	}
	return scanWithClamd(addr, data)
}

func scanWithClamd(addr string, data []byte) error {
	conn, err := net.DialTimeout("tcp", addr, clamTimeout)
	if err != nil {
		return fmt.Errorf("virus scanner unreachable: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(clamTimeout))

	// zINSTREAM: null-terminated command, then <4-byte big-endian length><bytes> chunks, then a zero length.
	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return err
	}
	var size [4]byte
	for off := 0; off < len(data); off += clamChunkSize {
		chunk := data[off:min(off+clamChunkSize, len(data))]
		binary.BigEndian.PutUint32(size[:], uint32(len(chunk)))
		if _, err := conn.Write(size[:]); err != nil {
			return err
		}
		if _, err := conn.Write(chunk); err != nil {
			return err
		}
	}
	binary.BigEndian.PutUint32(size[:], 0)
	if _, err := conn.Write(size[:]); err != nil {
		return err
	}

	reply, err := io.ReadAll(io.LimitReader(conn, 1024))
	if err != nil {
		return fmt.Errorf("virus scanner reply: %w", err)
	}
	answer := strings.TrimSpace(string(bytes.TrimRight(reply, "\x00")))
	switch {
	case strings.HasSuffix(answer, " OK"):
		return nil
	case strings.HasSuffix(answer, " FOUND"):
		signature := strings.TrimSuffix(strings.TrimPrefix(answer, "stream: "), " FOUND")
		return fmt.Errorf("%w: %s", ErrInfected, signature)
	default:
		// e.g. "INSTREAM size limit exceeded. ERROR"
		return fmt.Errorf("virus scanner error: %s", answer)
	}
}
