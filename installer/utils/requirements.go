package utils

import (
	"fmt"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

func CheckMem(size uint64) error {
	m := sigar.Mem{}
	err := m.Get()
	if err != nil {
		return err
	}

	total := m.Total / 1024 / 1024 / 1024
	if total < size-1 {
		return fmt.Errorf("your system does not have the minimum required memory: %v GB", size)
	}
	return nil
}

func CheckDisk(size uint64) error {
	d := sigar.FileSystemUsage{}
	err := d.Get("/")
	if err != nil {
		return err
	}

	free := d.Free / 1024 / 1024
	if free < size-1 {
		return fmt.Errorf("your system does not have the minimum required free disk space: %v GB", size)
	}

	return nil
}

func CheckCPU(cores int) error {
	c, _ := cpu.Counts(true)
	if c < cores {
		return fmt.Errorf("your system does not have the minimum required CPU cores: %v", cores)
	}
	return nil
}

func CheckDistro(distro string) error {
	info, _ := host.Info()
	if info.Platform != distro {
		return fmt.Errorf("your Linux distribution (%s) is not %s", info.Platform, distro)
	}
	return nil
}
