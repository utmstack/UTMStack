package updater

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

func DownloadFile(file File, url string, headers map[string]string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}

	destDir := file.DestinationPath
	destPath := filepath.Join(destDir, file.Name)

	err = utils.CreatePathIfNotExist(destDir)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	if utils.CheckIfPathExist(destPath) {
		if file.ReplacePrevious {
			err := os.Remove(destPath)
			if err != nil {
				return fmt.Errorf("error removing previous file: %v", err)
			}
		} else {
			config.Logger().Info("File already exists: %s\n", file.Name)
			return nil
		}
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer outFile.Close()

	hash := sha256.New()
	writer := io.MultiWriter(outFile, hash)

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving file: %v", err)
	}

	if file.Binary {
		if err := utils.RunCmd("chmod", "+x", "-R", filepath.Join(destPath)); err != nil {
			return err
		}
	}

	computedHash := fmt.Sprintf("%x", hash.Sum(nil))
	if computedHash != file.Hash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", file.Hash, hash)
	}

	config.Logger().Info("File downloaded successfully: %s\n", destPath)

	return nil
}
