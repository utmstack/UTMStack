# Build the UTMStack EDR AMSI provider DLL with MSVC.
#
# Requires: Visual Studio Build Tools + Windows SDK (amsi.h, amsi.lib) and the
# cross-compiler for the target arch. Build once per shipped arch: arm64 + amd64.
#
# Usage (from a normal PowerShell — no Developer prompt needed):
#   powershell -File build.ps1 -Arch arm64
#   powershell -File build.ps1 -Arch amd64
#
# Notes learned building this on Windows-on-ARM64 (2026-07):
#  - There is no native arm64-HOST arm64-TARGET cl.exe in a default install; use
#    the x64-hosted cross compiler (bin\Hostx64\<arch>\cl.exe), which runs under
#    x64 emulation on ARM64 Windows.
#  - vcvarsall.bat on an arm64 host does NOT fully populate INCLUDE/LIB for the
#    cross target (missing the MSVC CRT include → 'excpt.h' error), so we set
#    INCLUDE/LIB explicitly below.
param([ValidateSet('arm64','amd64')][string]$Arch = 'arm64')
$ErrorActionPreference = 'Stop'

$vsRoot = Get-ChildItem 'C:\Program Files (x86)\Microsoft Visual Studio\*\*\VC\Tools\MSVC' -Directory |
          Sort-Object FullName | Select-Object -Last 1
if (-not $vsRoot) { throw 'MSVC tools not found. Install VS Build Tools + Windows SDK.' }
$msvc = $vsRoot.FullName
$cl   = Join-Path $msvc "bin\Hostx64\$Arch\cl.exe"

$sdkInc = Get-ChildItem 'C:\Program Files (x86)\Windows Kits\10\Include' -Directory |
          Sort-Object Name | Select-Object -Last 1
$ver = $sdkInc.Name
$sdk = 'C:\Program Files (x86)\Windows Kits\10'

$env:INCLUDE = "$msvc\include;$sdk\Include\$ver\ucrt;$sdk\Include\$ver\shared;$sdk\Include\$ver\um;$sdk\Include\$ver\winrt"
$env:LIB     = "$msvc\lib\$Arch;$sdk\Lib\$ver\ucrt\$Arch;$sdk\Lib\$ver\um\$Arch"

Push-Location $PSScriptRoot
& $cl /nologo /LD /EHsc /O2 /DUNICODE /D_UNICODE /GS utmstack_amsi.cpp `
      /link /DEF:utmstack_amsi.def amsi.lib ole32.lib /OUT:utmstack_amsi.dll
Pop-Location
Write-Host "Built utmstack_amsi.dll for $Arch"
