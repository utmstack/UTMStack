package usecase

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

func writeFlowFile(path string, flow domain.Flow) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal([]domain.Flow{flow})
	if err != nil {
		return err
	}

	tf, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := tf.Name()
	defer func() { _ = os.Remove(tmp) }()

	if _, err := tf.Write(data); err != nil {
		_ = tf.Close()
		return err
	}
	if err := tf.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func readFlowFile(path string) (domain.Flow, error) {
	var flow domain.Flow
	data, err := os.ReadFile(path)
	if err != nil {
		return flow, err
	}
	var list []domain.Flow
	if err := yaml.Unmarshal(data, &list); err != nil {
		return flow, err
	}
	if len(list) == 0 {
		return flow, fmt.Errorf("flow file %s contains no flows", path)
	}
	return list[0], nil
}

func renameFlowFile(src, dst string) error {
	if src == dst {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func removeFlowFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
