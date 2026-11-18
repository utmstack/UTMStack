// UTMStack EDR — minimal AMSI provider (IAntimalwareProvider).
// Its only job: forward the content Windows submits (PowerShell/WSH/VBA/.NET)
// to the UTMStack EDR service over a named pipe and return the verdict. All
// engine/detection logic lives in the Go service; this DLL carries none, and
// no engine name appears in any of its strings (branding rule).
//
// Wire protocol (must match agent/edr/amsi/pipe_protocol.go):
//   request = LE32(length | 0x80000000), "UTA2", input-status byte,
//             UTF-8 appName, NUL, complete content bytes (if status == 0)
//   response = 0 clean, 1 detected, 2 incomplete/allow, 3 incomplete/block

#include <windows.h>
#include <amsi.h>
#include <aclapi.h>
#include <new>
#include "bounded_stream.h"

#pragma comment(lib, "advapi32.lib")

using utmstack_amsi::InputStatus;
static const ULONG kMaxScan = 50u * 1024u * 1024u;
static const ULONGLONG kRequestTimeoutMs = 35000;

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

class ScopedHandle {
    HANDLE h;
public:
    explicit ScopedHandle(HANDLE value) : h(value) {}
    ~ScopedHandle() { if (h && h != INVALID_HANDLE_VALUE) CloseHandle(h); }
    HANDLE get() const { return h; }
    ScopedHandle(const ScopedHandle&) = delete;
    ScopedHandle& operator=(const ScopedHandle&) = delete;
};

// Verify the connected kernel object, not its reusable name. A regular user
// cannot create a pipe owned by LocalSystem. READ_CONTROL is explicitly
// granted to authenticated clients so this does not require SeDebugPrivilege
// or permission to inspect the SYSTEM service's process token.
static bool IsTrustedServicePipe(HANDLE pipe) {
    PSID owner = nullptr;
    PSECURITY_DESCRIPTOR descriptor = nullptr;
    const DWORD error = GetSecurityInfo(pipe, SE_KERNEL_OBJECT, OWNER_SECURITY_INFORMATION,
                                       &owner, nullptr, nullptr, nullptr, &descriptor);
    if (error != ERROR_SUCCESS) { SetLastError(error); return false; }
    const bool trusted = owner && IsValidSid(owner) && IsWellKnownSid(owner, WinLocalSystemSid);
    if (descriptor) LocalFree(descriptor);
    if (!trusted) SetLastError(ERROR_ACCESS_DENIED);
    return trusted;
}

// All asynchronous buffers and OVERLAPPED objects stay alive until completion,
// including after a timeout races completion and CancelIoEx returns NOT_FOUND.
static bool PipeTransfer(HANDLE pipe, BYTE* data, DWORD size, bool write,
                         ULONGLONG deadline) {
    ScopedHandle event(CreateEventW(nullptr, TRUE, FALSE, nullptr));
    if (!event.get()) return false;
    DWORD offset = 0;
    while (offset < size) {
        const ULONGLONG now = GetTickCount64();
        if (now >= deadline) { SetLastError(ERROR_TIMEOUT); return false; }
        DWORD wanted = size - offset;
        if (wanted > 65536) wanted = 65536;
        OVERLAPPED ov = {};
        ov.hEvent = event.get();
        ResetEvent(event.get());
        DWORD done = 0;
        BOOL ok = write ? WriteFile(pipe, data + offset, wanted, nullptr, &ov)
                        : ReadFile(pipe, data + offset, wanted, nullptr, &ov);
        if (ok) ok = GetOverlappedResult(pipe, &ov, &done, FALSE);
        if (!ok && GetLastError() == ERROR_IO_PENDING) {
            const ULONGLONG waitNow = GetTickCount64();
            const ULONGLONG remaining = deadline > waitNow ? deadline - waitNow : 0;
            DWORD waitMs = remaining > 0xfffffffeULL ? 0xfffffffeUL : static_cast<DWORD>(remaining);
            const DWORD wait = WaitForSingleObject(event.get(), waitMs);
            if (wait != WAIT_OBJECT_0) {
                const DWORD failure = wait == WAIT_TIMEOUT ? ERROR_TIMEOUT : GetLastError();
                CancelIoEx(pipe, &ov);
                GetOverlappedResult(pipe, &ov, &done, TRUE);
                SetLastError(failure ? failure : ERROR_GEN_FAILURE);
                return false;
            }
            ok = GetOverlappedResult(pipe, &ov, &done, FALSE);
        }
        if (!ok || done == 0 || done > wanted) {
            if (ok) SetLastError(ERROR_BROKEN_PIPE);
            return false;
        }
        offset += done;
    }
    return true;
}

// Failures return a failed HRESULT. An unscanned prefix or unavailable service
// must never be represented as a successful NOT_DETECTED analysis.
static HRESULT ForwardToService(const wchar_t* appName, InputStatus status,
                                const BYTE* data, ULONG size,
                                ULONGLONG deadline, AMSI_RESULT* result) {
    char app[1024];
    int appLen = 0;
    if (appName) {
        appLen = WideCharToMultiByte(CP_UTF8, WC_ERR_INVALID_CHARS, appName, -1,
                                    app, sizeof(app), nullptr, nullptr);
        if (appLen > 0) appLen--;
        else { appLen = 0; status = InputStatus::InvalidAttributes; }
    }
    if (size > kMaxScan) status = InputStatus::SizeLimit;
    if (status != InputStatus::Complete) { data = nullptr; size = 0; }

    HANDLE pipe = INVALID_HANDLE_VALUE;
    for (int attempt = 0; attempt < 2; ++attempt) {
        if (GetTickCount64() >= deadline) return HRESULT_FROM_WIN32(ERROR_TIMEOUT);
        const DWORD clientRights = FILE_READ_DATA | FILE_WRITE_DATA | FILE_READ_ATTRIBUTES |
                                   READ_CONTROL | SYNCHRONIZE;
        pipe = CreateFileW(PIPE_NAME, clientRights, 0,
                           nullptr, OPEN_EXISTING, FILE_FLAG_OVERLAPPED |
                           SECURITY_SQOS_PRESENT | SECURITY_IDENTIFICATION, nullptr);
        if (pipe != INVALID_HANDLE_VALUE) break;
        if (GetLastError() != ERROR_PIPE_BUSY) return HRESULT_FROM_WIN32(GetLastError());
        const ULONGLONG now = GetTickCount64();
        if (now >= deadline) return HRESULT_FROM_WIN32(ERROR_TIMEOUT);
        DWORD waitMs = static_cast<DWORD>((deadline - now) > 2000 ? 2000 : deadline - now);
        if (!WaitNamedPipeW(PIPE_NAME, waitMs)) return HRESULT_FROM_WIN32(GetLastError());
    }
    if (pipe == INVALID_HANDLE_VALUE) return HRESULT_FROM_WIN32(ERROR_PIPE_BUSY);
    ScopedHandle connected(pipe);
    if (!IsTrustedServicePipe(pipe)) return HRESULT_FROM_WIN32(GetLastError());
    const DWORD total = static_cast<DWORD>(appLen) + 1 + size + 5;
    const DWORD length = total | 0x80000000UL;
    BYTE header[9] = { static_cast<BYTE>(length), static_cast<BYTE>(length >> 8),
                       static_cast<BYTE>(length >> 16), static_cast<BYTE>(length >> 24),
                       'U', 'T', 'A', '2', static_cast<BYTE>(status) };
    BYTE zero = 0;
    BYTE response = 0xff;
    if (!PipeTransfer(pipe, header, sizeof(header), true, deadline) ||
        (appLen > 0 && !PipeTransfer(pipe, reinterpret_cast<BYTE*>(app), appLen, true, deadline)) ||
        !PipeTransfer(pipe, &zero, 1, true, deadline) ||
        (size > 0 && !PipeTransfer(pipe, const_cast<BYTE*>(data), size, true, deadline)) ||
        !PipeTransfer(pipe, &response, 1, false, deadline)) {
        DWORD failure = GetLastError();
        return HRESULT_FROM_WIN32(failure ? failure : ERROR_GEN_FAILURE);
    }
    switch (response) {
    case 0:
        if (status != InputStatus::Complete) return HRESULT_FROM_WIN32(ERROR_INVALID_DATA);
        *result = AMSI_RESULT_NOT_DETECTED;
        return S_OK;
    case 1:
        *result = AMSI_RESULT_DETECTED;
        return S_OK;
    case 2:
        return HRESULT_FROM_WIN32(ERROR_NOT_SUPPORTED);
    case 3:
        *result = AMSI_RESULT_BLOCKED_BY_ADMIN_START;
        return S_OK;
    default:
        return HRESULT_FROM_WIN32(ERROR_INVALID_DATA);
    }
}

class Provider : public IAntimalwareProvider {
    LONG m_ref; // per-instance refcount (never shared with the factory)
public:
    Provider() : m_ref(1) { InterlockedIncrement(&g_locks); }
    ~Provider() { InterlockedDecrement(&g_locks); }

    STDMETHODIMP QueryInterface(REFIID riid, void** ppv) override {
        if (!ppv) return E_POINTER;
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
        if (!stream || !result) return E_INVALIDARG;
        const ULONGLONG deadline = GetTickCount64() + kRequestTimeoutMs;

        // IAmsiStream::GetAttribute(attribute, dataSize, data, retData).
        // CONTENT_SIZE is a ULONGLONG (8 bytes) and CONTENT_ADDRESS a PVOID
        // (8 bytes) — the buffers MUST match those widths or GetAttribute
        // overruns the stack (0xc0000409 /GS fast-fail).
        ULONGLONG size = 0;
        ULONG got = 0;
        HRESULT sizeHR = stream->GetAttribute(AMSI_ATTRIBUTE_CONTENT_SIZE,
                             sizeof(size), reinterpret_cast<unsigned char*>(&size), &got);
        const bool knownSize = SUCCEEDED(sizeHR) && got == sizeof(size);
        InputStatus status = InputStatus::Complete;
        if (SUCCEEDED(sizeHR) && !knownSize) status = InputStatus::InvalidAttributes;

        BYTE* addr = nullptr;
        got = 0;
        HRESULT addrHR = stream->GetAttribute(AMSI_ATTRIBUTE_CONTENT_ADDRESS,
                             sizeof(addr), reinterpret_cast<unsigned char*>(&addr), &got);
        if (FAILED(addrHR)) addr = nullptr;
        else if (got != sizeof(addr)) { addr = nullptr; status = InputStatus::InvalidAttributes; }

        wchar_t appName[256] = {};
        ULONG appGot = 0;
        HRESULT appHR = stream->GetAttribute(AMSI_ATTRIBUTE_APP_NAME,
                             sizeof(appName), reinterpret_cast<unsigned char*>(appName), &appGot);
        if (SUCCEEDED(appHR)) {
            bool terminated = false;
            if (appGot <= sizeof(appName) && appGot % sizeof(wchar_t) == 0) {
                for (ULONG i = 0; i < appGot / sizeof(wchar_t); ++i) {
                    if (appName[i] == 0) { terminated = true; break; }
                }
            }
            if (!terminated) { appName[0] = 0; status = InputStatus::InvalidAttributes; }
        } else {
            appName[0] = 0; // optional metadata; never use a partial attribute
        }
        if (knownSize && size > kMaxScan) status = InputStatus::SizeLimit;
        if (status != InputStatus::Complete) {
            return ForwardToService(appName, status, nullptr, 0, deadline, result);
        }
        if (knownSize && (addr || size == 0)) {
            return ForwardToService(appName, InputStatus::Complete, addr,
                                    static_cast<ULONG>(size), deadline, result);
        }
        auto input = utmstack_amsi::ReadCompleteStream(knownSize, size, kMaxScan,
            [stream](std::uint64_t position, std::uint32_t wanted, std::uint8_t* buffer, std::uint32_t& read) {
                ULONG count = 0;
                HRESULT hr = stream->Read(position, wanted, buffer, &count);
                read = count;
                return SUCCEEDED(hr);
            },
            [deadline]() { return GetTickCount64() >= deadline; });
        return ForwardToService(appName, input.status, input.data.get(), input.size, deadline, result);
    }

    STDMETHODIMP_(void) CloseSession(ULONGLONG) override {}

    STDMETHODIMP DisplayName(LPWSTR* displayName) override {
        if (!displayName) return E_POINTER;
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
        if (!ppv) return E_POINTER;
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
