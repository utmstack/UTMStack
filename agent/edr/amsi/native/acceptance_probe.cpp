// Harmless, disposable-Windows-VM acceptance client; never part of the provider.
#include <windows.h>
#include <amsi.h>
#include <objbase.h>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <cwchar>
#include <vector>
#include <string>
#include <algorithm>
#include <userenv.h>

// Disposable-lab helper: create a genuinely nonadmin process without depending
// on an interactive session or Task Scheduler demand-start policy. The random
// local-only password arrives over an anonymous stdin pipe, never argv/files.
static int RunUser(int argc,wchar_t** argv) {
    if(argc!=6)return 2;
    char secret[512]={};wchar_t password[512]={};
    if(!std::fgets(secret,sizeof(secret),stdin))return 2;
    size_t count=std::strlen(secret);
    while(count && (secret[count-1]=='\r'||secret[count-1]=='\n'))secret[--count]=0;
    if(!count || count>=511){SecureZeroMemory(secret,sizeof(secret));return 2;}
    for(size_t i=0;i<count;i++)password[i]=static_cast<unsigned char>(secret[i]);
    HANDLE token=nullptr;
    DWORD logonType=LOGON32_LOGON_BATCH;
    BOOL logged=LogonUserW(argv[2],L".",password,logonType,LOGON32_PROVIDER_DEFAULT,&token);
    if(!logged && GetLastError()==ERROR_LOGON_TYPE_NOT_GRANTED){
        // Stock nonadmin accounts need not hold SeBatchLogonRight. Use their
        // existing local interactive-logon right; never grant extra privileges.
        logonType=LOGON32_LOGON_INTERACTIVE;
        logged=LogonUserW(argv[2],L".",password,logonType,LOGON32_PROVIDER_DEFAULT,&token);
    }
    const DWORD logonError=logged?ERROR_SUCCESS:GetLastError();
    SecureZeroMemory(secret,sizeof(secret));SecureZeroMemory(password,sizeof(password));
    if(!logged){std::printf("FAIL LogonUser error=%lu\n",logonError);return 1;}
    std::printf("INFO standard-user logon_type=%lu\n",logonType);
    void* environment=nullptr;
    if(!CreateEnvironmentBlock(&environment,token,FALSE)){std::printf("FAIL environment error=%lu\n",GetLastError());CloseHandle(token);return 1;}
    HANDLE job=CreateJobObjectW(nullptr,nullptr);
    JOBOBJECT_EXTENDED_LIMIT_INFORMATION limits={};limits.BasicLimitInformation.LimitFlags=JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
    if(!job || !SetInformationJobObject(job,JobObjectExtendedLimitInformation,&limits,sizeof(limits))){std::printf("FAIL owned job error=%lu\n",GetLastError());if(job)CloseHandle(job);DestroyEnvironmentBlock(environment);CloseHandle(token);return 1;}
    const size_t commandLength=wcslen(argv[3]);
    if(!commandLength || commandLength>=32767){CloseHandle(job);DestroyEnvironmentBlock(environment);CloseHandle(token);return 2;}
    std::vector<wchar_t> command(argv[3],argv[3]+commandLength+1);
    STARTUPINFOW startup={};startup.cb=sizeof(startup);PROCESS_INFORMATION process={};
    BOOL created=CreateProcessAsUserW(token,nullptr,command.data(),nullptr,nullptr,FALSE,CREATE_SUSPENDED|CREATE_NO_WINDOW|CREATE_UNICODE_ENVIRONMENT,environment,argv[4],&startup,&process);
    const DWORD createError=created?ERROR_SUCCESS:GetLastError();
    DestroyEnvironmentBlock(environment);CloseHandle(token);
    if(!created){std::printf("FAIL CreateProcessAsUser error=%lu\n",createError);CloseHandle(job);return 1;}
    if(!AssignProcessToJobObject(job,process.hProcess) || ResumeThread(process.hThread)==static_cast<DWORD>(-1)){
        std::printf("FAIL owned job assignment/resume error=%lu\n",GetLastError());TerminateProcess(process.hProcess,1);CloseHandle(process.hThread);CloseHandle(process.hProcess);CloseHandle(job);return 1;
    }
    CloseHandle(process.hThread);
    const unsigned long requested=wcstoul(argv[5],nullptr,10);
    const DWORD timeout=requested>0 && requested<=180000?static_cast<DWORD>(requested):180000;
    const DWORD wait=WaitForSingleObject(process.hProcess,timeout);
    DWORD code=1;
    if(wait==WAIT_OBJECT_0)GetExitCodeProcess(process.hProcess,&code);
    else{std::printf("FAIL standard-user process wait=%lu\n",wait);TerminateJobObject(job,1);WaitForSingleObject(process.hProcess,5000);}
    CloseHandle(process.hProcess);CloseHandle(job);
    std::printf("%s standard-user-owned-process exit=%lu\n",code==0?"PASS":"FAIL",code);
    return code==0?0:1;
}

static const CLSID kProvider = {0x7b8e4f20,0x2c3d,0x4a5b,{0x9c,0x6e,0x1f,0x2a,0x3b,0x4c,0x5d,0x6e}};
static const char kMarker[] = "UTMSTACK_AMSI_ACCEPTANCE_2026";

class Stream final : public IAmsiStream {
    LONG refs = 1;
public:
    std::vector<BYTE> bytes;
    bool address = true, known = true, invalidSize = false, invalidApp = false;
    bool badRead = false;
    ULONG chunk = 65536;
    ULONGLONG claimed = 0;
    explicit Stream(const std::vector<BYTE>& content) : bytes(content), claimed(content.size()) {}
    HRESULT STDMETHODCALLTYPE QueryInterface(REFIID iid,void** out) override {
        if (!out) return E_POINTER;
        *out=nullptr;
        if (iid==IID_IUnknown || iid==__uuidof(IAmsiStream)) { *out=this; AddRef(); return S_OK; }
        return E_NOINTERFACE;
    }
    ULONG STDMETHODCALLTYPE AddRef() override {return InterlockedIncrement(&refs);}
    ULONG STDMETHODCALLTYPE Release() override {return InterlockedDecrement(&refs);}
    HRESULT STDMETHODCALLTYPE GetAttribute(AMSI_ATTRIBUTE attr,ULONG size,BYTE* out,ULONG* got) override {
        if (!got) return E_POINTER;
        *got=0;
        if (attr==AMSI_ATTRIBUTE_CONTENT_SIZE) {
            if (!known) return E_NOTIMPL;
            *got=invalidSize?4:sizeof(claimed);
            if (size<*got) return HRESULT_FROM_WIN32(ERROR_INSUFFICIENT_BUFFER);
            memcpy(out,&claimed,*got); return S_OK;
        }
        if (attr==AMSI_ATTRIBUTE_CONTENT_ADDRESS) {
            if (!address) return E_NOTIMPL;
            BYTE* pointer=bytes.empty()?nullptr:bytes.data(); *got=sizeof(pointer);
            if (size<*got) return HRESULT_FROM_WIN32(ERROR_INSUFFICIENT_BUFFER);
            memcpy(out,&pointer,*got); return S_OK;
        }
        if (attr==AMSI_ATTRIBUTE_APP_NAME) {
            static const wchar_t app[]=L"Inert acceptance \x2603";
            *got=sizeof(app)-(invalidApp?sizeof(wchar_t):0);
            if (size<*got) return HRESULT_FROM_WIN32(ERROR_INSUFFICIENT_BUFFER);
            memcpy(out,app,*got); return S_OK;
        }
        return E_NOTIMPL;
    }
    HRESULT STDMETHODCALLTYPE Read(ULONGLONG pos,ULONG size,BYTE* out,ULONG* got) override {
        if (!got) return E_POINTER;
        if (badRead) { *got=size+1; return S_OK; }
        *got=0; if (pos>=bytes.size()) return S_OK;
        *got=(std::min)(size,(std::min)(chunk,static_cast<ULONG>(bytes.size()-static_cast<size_t>(pos))));
        memcpy(out,bytes.data()+static_cast<size_t>(pos),*got); return S_OK;
    }
};

static int Check(const char* name,HRESULT hr,AMSI_RESULT result,bool incomplete,bool detected,bool closed) {
    const bool ok=incomplete?(closed?(SUCCEEDED(hr) && result>=AMSI_RESULT_BLOCKED_BY_ADMIN_START):FAILED(hr)):
        SUCCEEDED(hr) && (detected?(result>=AMSI_RESULT_DETECTED):(result==AMSI_RESULT_NOT_DETECTED));
    std::printf("%s %s hr=0x%08lx result=%u\n",ok?"PASS":"FAIL",name,static_cast<unsigned long>(hr),static_cast<unsigned>(result));
    return ok?0:1;
}

int wmain(int argc,wchar_t** argv) {
    if(argc>1 && wcscmp(argv[1],L"run-user")==0)return RunUser(argc,argv);
    if (argc!=3) { std::fprintf(stderr,"usage: acceptance_probe DLL open|closed|transport-failure|registered|registered-owned|registered-owned-official\n"); return 2; }
    HANDLE identity=nullptr; TOKEN_ELEVATION elevation={}; DWORD identityBytes=0;
    if(!OpenProcessToken(GetCurrentProcess(),TOKEN_QUERY,&identity))return 1;
    BYTE identityBuffer[512]={};
    const bool identityOK=GetTokenInformation(identity,TokenElevation,&elevation,sizeof(elevation),&identityBytes) && GetTokenInformation(identity,TokenUser,identityBuffer,sizeof(identityBuffer),&identityBytes);
    const bool system=identityOK && IsWellKnownSid(reinterpret_cast<TOKEN_USER*>(identityBuffer)->User.Sid,WinLocalSystemSid);
    CloseHandle(identity);
    if(!identityOK)return 1;
    std::printf("INFO native-identity elevated=%s system=%s\n",elevation.TokenIsElevated?"true":"false",system?"true":"false");
    const bool closed=wcscmp(argv[2],L"closed")==0;
    const bool official=wcscmp(argv[2],L"registered-owned-official")==0;
    if (wcscmp(argv[2],L"registered-owned")==0 || official) {
        HRESULT hr=CoInitializeEx(nullptr,COINIT_MULTITHREADED);
        if(FAILED(hr)){std::printf("FAIL owned CoInitializeEx hr=0x%08lx\n",static_cast<unsigned long>(hr));return 1;}
        IAntimalwareProvider* provider=nullptr;
        hr=CoCreateInstance(kProvider,nullptr,CLSCTX_INPROC_SERVER,__uuidof(IAntimalwareProvider),reinterpret_cast<void**>(&provider));
        if(FAILED(hr)||!provider){std::printf("FAIL owned registered class hr=0x%08lx\n",static_cast<unsigned long>(hr));CoUninitialize();return 1;}
        std::vector<wchar_t> actual(32768),expected(32768);
        HMODULE module=GetModuleHandleW(L"utmstack_amsi.dll");
        DWORD got=module?GetModuleFileNameW(module,actual.data(),static_cast<DWORD>(actual.size())):0;
        DWORD wanted=GetFullPathNameW(argv[1],static_cast<DWORD>(expected.size()),expected.data(),nullptr);
        const bool owned=got>0 && got<actual.size() && wanted>0 && wanted<expected.size() && _wcsicmp(actual.data(),expected.data())==0;
        std::printf("%s registered-owned-module %ls\n",owned?"PASS":"FAIL",got?actual.data():L"unavailable");
        int failures=owned?0:1;
        if(owned){
            std::vector<BYTE> clean{'i','n','e','r','t',' ','c','l','e','a','n'};
            std::vector<BYTE> marker(kMarker,kMarker+sizeof(kMarker)-1);
            if(official){
                // Public inert EICAR data exists only in the scan buffer.
                static const BYTE encoded[]={2,111,21,123,10,127,26,27,10,1,110,6,10,0,2,111,110,114,10,4,115,109,25,25,115,109,39,126,31,19,25,27,8,119,9,14,27,20,30,27,8,30,119,27,20,14,19,12,19,8,15,9,119,14,31,9,14,119,28,19,22,31,123,126,18,113,18,112};
                volatile BYTE key=0x5a;marker.clear();
                for(BYTE byte:encoded)marker.push_back(byte^key);
            }
            Stream a(clean),b(marker);AMSI_RESULT result=AMSI_RESULT_CLEAN;
            hr=provider->Scan(&a,&result);failures+=Check("registered-owned-clean",hr,result,false,false,false);
            result=AMSI_RESULT_CLEAN;hr=provider->Scan(&b,&result);failures+=Check(official?"registered-owned-official-eicar":"registered-owned-marker",hr,result,false,true,false);
        }
        provider->Release();CoUninitialize();return failures?1:0;
    }
    if (wcscmp(argv[2],L"registered")==0) {
        HAMSICONTEXT context=nullptr; HRESULT hr=AmsiInitialize(L"UTMStack inert acceptance",&context);
        if (FAILED(hr)) {std::printf("FAIL AmsiInitialize hr=0x%08lx\n",static_cast<unsigned long>(hr));return 1;}
        AMSI_RESULT result=AMSI_RESULT_CLEAN;
        hr=AmsiScanBuffer(context,const_cast<char*>(kMarker),static_cast<ULONG>(sizeof(kMarker)-1),L"inert-marker",nullptr,&result);
        const bool ok=SUCCEEDED(hr) && AmsiResultIsMalware(result);
        std::printf("%s registered-AMSI-marker hr=0x%08lx result=%u\n",ok?"PASS":"FAIL",static_cast<unsigned long>(hr),static_cast<unsigned>(result));
        AmsiUninitialize(context); return ok?0:1;
    }
    HMODULE dll=LoadLibraryW(argv[1]);
    if (!dll) {std::printf("FAIL LoadLibrary error=%lu\n",GetLastError());return 1;}
    using GetFactory=HRESULT (STDAPICALLTYPE *)(REFCLSID,REFIID,void**);
    auto getFactory=reinterpret_cast<GetFactory>(GetProcAddress(dll,"DllGetClassObject"));
    if (!getFactory) {std::puts("FAIL DllGetClassObject missing");return 1;}
    IClassFactory* factory=nullptr;
    HRESULT hr=getFactory(kProvider,IID_IClassFactory,reinterpret_cast<void**>(&factory));
    if (FAILED(hr)||!factory) {std::puts("FAIL class factory");return 1;}
    IAntimalwareProvider* provider=nullptr;
    hr=factory->CreateInstance(nullptr,__uuidof(IAntimalwareProvider),reinterpret_cast<void**>(&provider));
    factory->Release();
    if (FAILED(hr)||!provider) {std::puts("FAIL provider instance");return 1;}
    int failures=0;
    std::vector<BYTE> clean{'i','n','e','r','t',' ','c','l','e','a','n'};
    std::vector<BYTE> marker(kMarker,kMarker+sizeof(kMarker)-1);
    auto scan=[&](const char* name,Stream& stream,bool incomplete,bool detected) {
        AMSI_RESULT result=AMSI_RESULT_CLEAN;
        HRESULT scanHR=provider->Scan(&stream,&result);
        failures+=Check(name,scanHR,result,incomplete,detected,closed);
    };
    if (wcscmp(argv[2],L"transport-failure")==0) {
        Stream stream(clean); AMSI_RESULT result=AMSI_RESULT_CLEAN;
        hr=provider->Scan(&stream,&result);
        const bool ok=FAILED(hr);
        std::printf("%s transport-failure hr=0x%08lx result=%u\n",ok?"PASS":"FAIL",static_cast<unsigned long>(hr),static_cast<unsigned>(result));
        provider->Release();FreeLibrary(dll);return ok?0:1;
    }
    Stream a(clean);scan("address-clean",a,false,false);
    Stream b(marker);scan("address-marker",b,false,true);
    std::vector<BYTE> wide;
    for (BYTE value:marker) {wide.push_back(value);wide.push_back(0);}
    Stream wideStream(wide);scan("raw-UTF16LE-address-marker",wideStream,false,true);
    std::vector<BYTE> tail(150000,'x');tail.insert(tail.end(),marker.begin(),marker.end());
    Stream c(tail);c.address=false;scan("known-stream-tail-after-64KiB",c,false,true);
    Stream d(tail);d.address=false;d.known=false;scan("unknown-stream-tail-after-64KiB",d,false,true);
    Stream e(tail);e.address=false;e.chunk=113;scan("short-successful-stream-reads",e,false,true);
    Stream f(clean);f.claimed=50ULL*1024*1024+1;scan("oversized-address-never-clean",f,true,false);
    Stream g(clean);g.invalidSize=true;scan("invalid-attribute-width",g,true,false);
    Stream h(clean);h.invalidApp=true;scan("unterminated-app-name",h,true,false);
    Stream i(clean);i.address=false;i.claimed++;scan("premature-EOF-never-clean",i,true,false);
    Stream j(clean);j.address=false;j.badRead=true;scan("invalid-read-count",j,true,false);
    Stream k(std::vector<BYTE>{});scan("empty-address",k,false,false);
    provider->Release();FreeLibrary(dll);return failures?1:0;
}
