package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCgroupGoMemoryTarget(t *testing.T) {
	for _, test := range []struct {
		limit uint64
		want  int64
	}{
		{512 * goMemoryMiB, 0},
		{goMemoryGiB, 768 << 20},
		{1280 * goMemoryMiB, 960 << 20},
		{2 * goMemoryGiB, 1000 << 20},
		{3 * goMemoryGiB, 0},
	} {
		if got := cgroupGoMemoryTarget(test.limit); got != test.want {
			t.Errorf("limit %d: got %d, want %d", test.limit, got, test.want)
		}
	}
}

func TestMemoryHierarchyUsesCurrentAndParentBudgets(t *testing.T) {
	for _, test := range []struct {
		name, membership, mountRoot, filesystem string
		files                                   map[string]string
		want                                    uint64
	}{
		{"systemd unit", "0::/system.slice/UTMStackEDR.service\n", "/", "cgroup2", map[string]string{"memory.max": "max", "system.slice/UTMStackEDR.service/memory.max": "2147483648"}, 2 * goMemoryGiB},
		{"parent tighter than leaf", "0::/parent/service\n", "/", "cgroup2", map[string]string{"memory.max": "4294967296", "parent/memory.max": "1073741824", "parent/service/memory.max": "2147483648"}, goMemoryGiB},
		{"unlimited leaf inherits parent", "0::/parent/service\n", "/", "cgroup2", map[string]string{"memory.max": "max", "parent/memory.max": "2147483648", "parent/service/memory.max": "max"}, 2 * goMemoryGiB},
		{"root tighter than descendants", "0::/parent/service\n", "/", "cgroup2", map[string]string{"memory.max": "1073741824", "parent/service/memory.max": "2147483648"}, goMemoryGiB},
		{"private namespace root", "0::/\n", "/", "cgroup2", map[string]string{"memory.max": "2147483648"}, 2 * goMemoryGiB},
		{"remapped mount", "0::/host/container/service\n", "/host/container", "cgroup2", map[string]string{"memory.max": "2147483648", "service/memory.max": "1073741824"}, goMemoryGiB},
		{"escaped mount path", "0::/host/group space/service\n", "/host/group\\040space", "cgroup2", map[string]string{"service/memory.max": "1073741824"}, goMemoryGiB},
		{"outside mounted subtree", "0::/host/foreign/service\n", "/host/container", "cgroup2", map[string]string{"host/foreign/service/memory.max": "1073741824"}, 0},
		{"v1 memory in controller list", "4:cpu,memory:/parent/service\n", "/", "cgroup", map[string]string{"memory/memory.limit_in_bytes": "9223372036854771712", "memory/parent/memory.limit_in_bytes": "1610612736", "memory/parent/service/memory.limit_in_bytes": "2147483648"}, 1536 * goMemoryMiB},
		{"v1 remapped mount", "4:memory:/host/container/service\n", "/host/container", "cgroup", map[string]string{"memory/memory.limit_in_bytes": "2147483648", "memory/service/memory.limit_in_bytes": "1073741824"}, goMemoryGiB},
		{"other controller ignored", "3:cpu:/parent/service\n", "/", "cgroup", map[string]string{"memory/parent/service/memory.limit_in_bytes": "1073741824"}, 0},
		{"invalid and unlimited values ignored", "0::/parent/service\n", "/", "cgroup2", map[string]string{"memory.max": "max", "parent/memory.max": "-1", "parent/service/memory.max": "0"}, 0},
		{"traversal rejected", "0::/parent/../foreign\n", "/", "cgroup2", map[string]string{"foreign/memory.max": "1073741824"}, 0},
		{"relative membership rejected", "0::parent/service\n", "/", "cgroup2", map[string]string{"parent/service/memory.max": "1073741824"}, 0},
		{"path NUL rejected", "0::/parent/\x00\n", "/", "cgroup2", map[string]string{"memory.max": "max"}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			for name, value := range test.files {
				p := filepath.Join(root, name)
				if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(value+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			point := root
			if test.filesystem == "cgroup" {
				point = filepath.Join(root, "memory")
			}
			mounts := fmt.Sprintf("31 20 0:29 %s %s rw,nosuid,nodev - %s cgroup rw\n", test.mountRoot, point, test.filesystem)
			if got := readCgroupMemoryHierarchy(root, test.membership, mounts); got != test.want {
				t.Fatalf("effective memory cap=%d want %d", got, test.want)
			}
		})
	}
}

func TestHierarchyPathDoesNotTraverseOutsideMount(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(root, "foreign")
	if err := os.Mkdir(outside, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "memory.max"), []byte("1073741824"), 0600); err != nil {
		t.Fatal(err)
	}
	mount := filepath.Join(root, "mounted")
	if err := os.Mkdir(mount, 0700); err != nil {
		t.Fatal(err)
	}
	if got := readCgroupMemoryHierarchy(mount, "0::/../foreign\n", ""); got != 0 {
		t.Fatalf("adopted outside budget %d", got)
	}
	if got := readCgroupMemoryHierarchy("relative", "0::/\n", ""); got != 0 {
		t.Fatalf("adopted relative mount %d", got)
	}
}

func TestCgroupMountEscapesAndMalformedFields(t *testing.T) {
	for _, encoded := range []string{"\\", "\\04", "\\999", "\\000"} {
		if _, ok := decodeCgroupMountPath(encoded); ok {
			t.Fatalf("accepted invalid mount escape %q", encoded)
		}
	}
	if got, ok := decodeCgroupMountPath("/space\\040tab\\011slash\\134"); !ok || got != "/space tab\tslash\\" {
		t.Fatalf("mount escaping: %q %v", got, ok)
	}
	if got := cgroupMountRoot("/sys/fs/cgroup", "cgroup2", "1 2 0:3 / /sys/fs/cgroup rw -\n"); got != "/" {
		t.Fatal(got)
	}
	if validCgroupPath("/" + strings.Repeat("a", 4096)) {
		t.Fatal("accepted oversized cgroup path")
	}
}

func TestDesiredGoMemoryLimitPreservesExplicitSetting(t *testing.T) {
	if got := desiredGoMemoryLimit("900MiB", goMemoryGiB); got != 0 {
		t.Fatalf("explicit target overwritten: %d", got)
	}
	if got := desiredGoMemoryLimit("", goMemoryGiB); got != 768<<20 {
		t.Fatalf("missing automatic target: %d", got)
	}
}

func TestReadCgroupMemoryLimit(t *testing.T) {
	root := t.TempDir()
	write := func(name, value string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("memory.max", "max\n")
	write("memory/memory.limit_in_bytes", "1073741824\n")
	if got := readCgroupMemoryLimit(root); got != goMemoryGiB {
		t.Fatalf("v1 fallback: %d", got)
	}
	write("memory.max", "2147483648\n")
	if got := readCgroupMemoryLimit(root); got != goMemoryGiB {
		t.Fatalf("smaller limit: %d", got)
	}
	write("memory/memory.limit_in_bytes", "0\n")
	if got := readCgroupMemoryLimit(root); got != 2*goMemoryGiB {
		t.Fatalf("v2 limit: %d", got)
	}
}
