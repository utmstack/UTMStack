package engine

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

const dialTimeout = 5 * time.Second

// ScanBytes streams data to the scan engine over the INSTREAM protocol and
// returns the verdict. NOTE: "clamd" here is an internal identifier only;
// nothing from this function is surfaced in events or logs (branding rules).
func ScanBytes(addr string, data []byte) (bool, string, error) {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return false, "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))

	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return false, "", err
	}
	const chunk = 8192
	for i := 0; i < len(data); i += chunk {
		end := i + chunk
		if end > len(data) {
			end = len(data)
		}
		var sz [4]byte
		binary.BigEndian.PutUint32(sz[:], uint32(end-i))
		if _, err := conn.Write(sz[:]); err != nil {
			return false, "", err
		}
		if _, err := conn.Write(data[i:end]); err != nil {
			return false, "", err
		}
	}
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return false, "", err
	}

	resp, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return false, "", err
	}
	resp = strings.TrimRight(resp, "\x00\n ")
	if strings.HasSuffix(resp, "OK") {
		return true, "", nil
	}
	if strings.HasSuffix(resp, "FOUND") {
		// "stream: <signature> FOUND"
		s := strings.TrimPrefix(resp, "stream: ")
		s = strings.TrimSuffix(s, " FOUND")
		return false, s, nil
	}
	return false, "", fmt.Errorf("unexpected engine response: %q", resp)
}

func Ping(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(dialTimeout))
	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return err
	}
	resp, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.TrimSpace(resp), "PONG") {
		return fmt.Errorf("unexpected ping response: %q", resp)
	}
	return nil
}

func Version(addr string) (string, error) {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(dialTimeout))
	if _, err := conn.Write([]byte("zVERSION\x00")); err != nil {
		return "", err
	}
	resp, err := bufio.NewReader(conn).ReadString(0)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(resp, "\x00\n"), nil
}

// SigDBVersion returns just the signature-database version from the engine's
// VERSION banner ("<engine> <engine-ver>/<sigdb-ver>/<date>"), dropping the
// engine-name field so nothing engine-branded is ever surfaced. Returns "" if
// the engine is unreachable or the banner is not in the expected form.
func SigDBVersion(addr string) string {
	raw, err := Version(addr)
	if err != nil {
		return ""
	}
	parts := strings.Split(raw, "/")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
