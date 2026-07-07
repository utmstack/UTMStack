// UTMStack EDR — minimal AMSI provider (IAntimalwareProvider).
// Its only job: forward the content Windows submits (PowerShell/WSH/VBA/.NET)
// to the UTMStack EDR service over a named pipe and return the verdict. All
// engine/detection logic lives in the Go service; this DLL carries none, and
// no engine name appears in any of its strings (branding rule).
//
// Wire protocol (must match agent/edr/amsi/pipe_windows.go):
//   request  = 4-byte LE length + "appName\0" + content bytes
//   response = 1 byte: 0 = allow, 1 = block

#include <windows.h>
#include <amsi.h>
#include <new>

// {7B8E4F20-2C3D-4A5B-9C6E-1F2A3B4C5D6E} — MUST match register_windows.go CLSID.
static const CLSID CLSID_UtmStackAmsi =
    { 0x7b8e4f20, 0x2c3d, 0x4a5b, { 0x9c, 0x6e, 0x1f, 0x2a, 0x3b, 0x4c, 0x5d, 0x6e } };

static const wchar_t* PIPE_NAME = L"\\\\.\\pipe\\utmstack_edr_amsi";

// Number of live provider objects + outstanding server locks. This — NOT a
// refcount shared with the class factory — gates DllCanUnloadNow. The factory is
// a static singleton the host releases right after CreateInstance (standard COM);
// if its Release could drive the module count to zero, the DLL would be unloaded
// while the host still holds a live provider, and the host's next call into the
// (now-unmapped) provider faults with 0xc0000409. Keep provider lifetime and
// factory lifetime on separate counters.
static LONG g_locks = 0;

// Returns true to BLOCK (malicious). Fails open (false) on any transport error.
static bool ForwardToService(const wchar_t* appName, const BYTE* data, ULONG size) {
    HANDLE pipe = INVALID_HANDLE_VALUE;
    for (int attempt = 0; attempt < 2; ++attempt) {
        pipe = CreateFileW(PIPE_NAME, GENERIC_READ | GENERIC_WRITE, 0,
                           nullptr, OPEN_EXISTING, 0, nullptr);
        if (pipe != INVALID_HANDLE_VALUE) break;
        if (GetLastError() != ERROR_PIPE_BUSY) return false;
        if (!WaitNamedPipeW(PIPE_NAME, 2000)) return false;
    }
    if (pipe == INVALID_HANDLE_VALUE) return false;

    char app[256];
    int appLen = 0;
    if (appName) {
        appLen = WideCharToMultiByte(CP_UTF8, 0, appName, -1, app, sizeof(app) - 1, nullptr, nullptr);
        if (appLen > 0) appLen--; // drop the counted null; we write it explicitly
    }

    DWORD total = (DWORD)appLen + 1 + size;
    DWORD written = 0;
    bool ok = WriteFile(pipe, &total, 4, &written, nullptr) != 0;
    if (ok && appLen > 0) ok = WriteFile(pipe, app, appLen, &written, nullptr) != 0;
    char zero = 0;
    if (ok) ok = WriteFile(pipe, &zero, 1, &written, nullptr) != 0;
    if (ok && size > 0) ok = WriteFile(pipe, data, size, &written, nullptr) != 0;

    BYTE resp = 0;
    DWORD read = 0;
    bool got = ok && ReadFile(pipe, &resp, 1, &read, nullptr) && read == 1;
    CloseHandle(pipe);
    return got && resp == 1;
}

class Provider : public IAntimalwareProvider {
    LONG m_ref; // per-instance refcount (never shared with the factory)
public:
    Provider() : m_ref(1) { InterlockedIncrement(&g_locks); }
    ~Provider() { InterlockedDecrement(&g_locks); }

    STDMETHODIMP QueryInterface(REFIID riid, void** ppv) override {
        if (riid == IID_IUnknown || riid == __uuidof(IAntimalwareProvider)) {
            *ppv = static_cast<IAntimalwareProvider*>(this);
            AddRef();
            return S_OK;
        }
        *ppv = nullptr;
        return E_NOINTERFACE;
    }
    STDMETHODIMP_(ULONG) AddRef() override { return InterlockedIncrement(&m_ref); }
    STDMETHODIMP_(ULONG) Release() override {
        LONG r = InterlockedDecrement(&m_ref);
        if (r == 0) delete this;
        return r;
    }

    STDMETHODIMP Scan(IAmsiStream* stream, AMSI_RESULT* result) override {
        if (result) *result = AMSI_RESULT_NOT_DETECTED;
        if (!stream || !result) return S_OK;

        // IAmsiStream::GetAttribute(attribute, dataSize, data, retData).
        // CONTENT_SIZE is a ULONGLONG (8 bytes) and CONTENT_ADDRESS a PVOID
        // (8 bytes) — the buffers MUST match those widths or GetAttribute
        // overruns the stack (0xc0000409 /GS fast-fail).
        ULONGLONG size = 0;
        ULONG got = 0;
        stream->GetAttribute(AMSI_ATTRIBUTE_CONTENT_SIZE,
                             sizeof(size), (unsigned char*)&size, &got);

        BYTE* addr = nullptr;
        stream->GetAttribute(AMSI_ATTRIBUTE_CONTENT_ADDRESS,
                             sizeof(addr), (unsigned char*)&addr, &got);

        wchar_t appName[256];
        appName[0] = 0;
        ULONG appGot = 0;
        stream->GetAttribute(AMSI_ATTRIBUTE_APP_NAME,
                             sizeof(appName) - sizeof(wchar_t), (unsigned char*)appName, &appGot);
        appName[255] = 0; // guarantee null-termination for WideCharToMultiByte

        const ULONG kMaxScan = 50u * 1024u * 1024u;
        bool block = false;
        if (addr && size > 0) {
            ULONG n = (size > kMaxScan) ? kMaxScan : (ULONG)size;
            block = ForwardToService(appName, addr, n);
        } else {
            // IAmsiStream::Read(position, size, buffer, readSize)
            BYTE buf[65536];
            ULONG readSize = 0;
            if (SUCCEEDED(stream->Read(0, sizeof(buf), buf, &readSize)) && readSize > 0) {
                block = ForwardToService(appName, buf, readSize);
            }
        }
        if (block) *result = AMSI_RESULT_DETECTED;
        return S_OK;
    }

    STDMETHODIMP_(void) CloseSession(ULONGLONG) override {}

    STDMETHODIMP DisplayName(LPWSTR* displayName) override {
        static const wchar_t name[] = L"UTMStack EDR";
        *displayName = (LPWSTR)CoTaskMemAlloc(sizeof(name));
        if (!*displayName) return E_OUTOFMEMORY;
        memcpy(*displayName, name, sizeof(name));
        return S_OK;
    }
};

class Factory : public IClassFactory {
public:
    STDMETHODIMP QueryInterface(REFIID riid, void** ppv) override {
        if (riid == IID_IUnknown || riid == IID_IClassFactory) {
            *ppv = static_cast<IClassFactory*>(this);
            AddRef();
            return S_OK;
        }
        *ppv = nullptr;
        return E_NOINTERFACE;
    }
    // Static singleton: fixed refcounts, never freed. LockServer adjusts the
    // module lock count so an outstanding host lock keeps the DLL loaded.
    STDMETHODIMP_(ULONG) AddRef() override { return 2; }
    STDMETHODIMP_(ULONG) Release() override { return 1; }
    STDMETHODIMP CreateInstance(IUnknown* outer, REFIID riid, void** ppv) override {
        if (outer) return CLASS_E_NOAGGREGATION;
        Provider* p = new (std::nothrow) Provider();
        if (!p) return E_OUTOFMEMORY;
        HRESULT hr = p->QueryInterface(riid, ppv);
        p->Release();
        return hr;
    }
    STDMETHODIMP LockServer(BOOL lock) override {
        if (lock) InterlockedIncrement(&g_locks);
        else InterlockedDecrement(&g_locks);
        return S_OK;
    }
};

static Factory g_factory;

STDAPI DllGetClassObject(REFCLSID rclsid, REFIID riid, void** ppv) {
    if (rclsid == CLSID_UtmStackAmsi) return g_factory.QueryInterface(riid, ppv);
    return CLASS_E_CLASSNOTAVAILABLE;
}
STDAPI DllCanUnloadNow() { return g_locks == 0 ? S_OK : S_FALSE; }
BOOL WINAPI DllMain(HINSTANCE h, DWORD reason, LPVOID) {
    if (reason == DLL_PROCESS_ATTACH) DisableThreadLibraryCalls(h);
    return TRUE;
}
