package dependency

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/utmstack/UTMStack/agent/edr/release"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const maxEDRReleaseDownload = 290 << 20

func newEDRReleaseClient(skipCertValidation bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertValidation}},
		Timeout:   5 * time.Minute,
	}
}

// download uses an exclusive file in a private, per-reconciliation directory.
// The shared generic downloader uses os.Create on a stable name and has no
// byte limit, so it is unsuitable for executable release archives.
func (u *releaseEDRUpdate) download(server string, skipCertValidation bool) (err error) {
	if _, err := u.key(); err != nil {
		return err
	}
	if err := os.MkdirAll(u.base, 0755); err != nil {
		return err
	}
	dir, err := release.MakePrivateTempDir(u.base, ".edr-download-")
	if err != nil {
		return err
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(dir)
		}
	}()
	client := u.client(skipCertValidation)
	defer client.CloseIdleConnections()
	request, err := http.NewRequest(http.MethodGet, u.url(server), nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("native EDR release download returned %s", response.Status)
	}
	if response.ContentLength > maxEDRReleaseDownload {
		return errors.New("native EDR release exceeds download byte limit")
	}
	path := filepath.Join(dir, "release.zip")
	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	n, copyErr := io.Copy(output, io.LimitReader(response.Body, maxEDRReleaseDownload+1))
	syncErr := output.Sync()
	closeErr := output.Close()
	if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
		return err
	}
	if n < 1 || n > maxEDRReleaseDownload {
		return errors.New("native EDR release exceeds download byte limit")
	}
	u.downloadDir, u.archive = dir, path
	complete = true
	return nil
}
