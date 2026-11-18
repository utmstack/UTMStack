# Build the AMSI provider with installed MSVC tools and Windows SDK libraries.
# Examples: powershell -File build.ps1 -Arch amd64
#           powershell -File build.ps1 -Arch arm64
# Optional paths/versions pin the build without relying on a Developer prompt.
[CmdletBinding()]
param(
    [ValidateSet('arm64', 'amd64')][string]$Arch = 'arm64',
    [string]$VisualStudioPath,
    [string]$WindowsSdkPath,
    [string]$ToolsetVersion,
    [string]$SdkVersion,
    [switch]$BuildAcceptanceProbe
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'The native AMSI build requires Windows, MSVC and the Windows SDK.' }
$target = (@{ amd64 = 'x64'; arm64 = 'arm64' })[$Arch]

function Get-VersionDirectories([string]$ParentPath, [string]$PinnedVersion) {
    if (-not (Test-Path -LiteralPath $ParentPath -PathType Container)) { return }
    Get-ChildItem -LiteralPath $ParentPath -Directory | Where-Object {
        $parsedVersion = $null
        [version]::TryParse($_.Name, [ref]$parsedVersion) -and
            (-not $PinnedVersion -or $_.Name -eq $PinnedVersion)
    } | Sort-Object { [version]$_.Name } -Descending
}

function Test-RequiredFiles([string]$BasePath, [string[]]$RelativeFiles) {
    foreach ($relativeFile in $RelativeFiles) {
        if (-not (Test-Path -LiteralPath (Join-Path $BasePath $relativeFile) -PathType Leaf)) { return $false }
    }
    return $true
}

$programRoots = @(${env:ProgramFiles(x86)}, $env:ProgramFiles) | Where-Object { $_ } | Select-Object -Unique
$installations = @()
if ($VisualStudioPath) {
    $installations = @($VisualStudioPath)
} else {
    if ($env:VSINSTALLDIR) { $installations += $env:VSINSTALLDIR }
    $vswhere = $null
    foreach ($programRoot in $programRoots) {
        $candidate = Join-Path $programRoot 'Microsoft Visual Studio\Installer\vswhere.exe'
        if (Test-Path -LiteralPath $candidate -PathType Leaf) { $vswhere = $candidate; break }
    }
    if (-not $vswhere) {
        $command = Get-Command vswhere.exe -ErrorAction SilentlyContinue
        if ($command) { $vswhere = $command.Source }
    }
    if ($vswhere) {
        $found = @(& $vswhere -all -products '*' -sort -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath)
        if ($LASTEXITCODE -ne 0) { throw "Visual Studio discovery failed (vswhere exit $LASTEXITCODE)." }
        $installations += $found | ForEach-Object { $_.Trim() } | Where-Object { $_ }
    }
    # Fallback for installations without vswhere, including VS Build Tools.
    foreach ($programRoot in $programRoots) {
        $pattern = Join-Path $programRoot 'Microsoft Visual Studio\*\*'
        $installations += @(Get-ChildItem -Path $pattern -Directory -ErrorAction SilentlyContinue | Select-Object -ExpandProperty FullName)
    }
}

$msvc = $null
foreach ($installation in ($installations | Select-Object -Unique)) {
    $toolsRoot = Join-Path $installation 'VC\Tools\MSVC'
    foreach ($candidate in @(Get-VersionDirectories $toolsRoot $ToolsetVersion)) {
        if (Test-RequiredFiles $candidate.FullName @("bin\Hostx64\$target\cl.exe", "bin\Hostx64\$target\link.exe", 'include\vcruntime.h', "lib\$target\libcmt.lib")) {
            $msvc = $candidate.FullName
            break
        }
    }
    if ($msvc) { break }
}
if (-not $msvc) { throw "MSVC x64-hosted $target tools/libraries were not found. Install the target C++ build tools, or specify -VisualStudioPath and -ToolsetVersion." }

$sdkRoots = @()
if ($WindowsSdkPath) {
    $sdkRoots = @($WindowsSdkPath)
} else {
    if ($env:WindowsSdkDir) { $sdkRoots += $env:WindowsSdkDir }
    foreach ($registryPath in @('HKLM:\SOFTWARE\Microsoft\Windows Kits\Installed Roots', 'HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows Kits\Installed Roots')) {
        $kits = Get-ItemProperty -LiteralPath $registryPath -Name KitsRoot10 -ErrorAction SilentlyContinue
        if ($kits) { $sdkRoots += $kits.KitsRoot10 }
    }
    foreach ($programRoot in $programRoots) { $sdkRoots += Join-Path $programRoot 'Windows Kits\10' }
}
$sdk = $null
$sdkVersionFound = $null
foreach ($sdkRoot in ($sdkRoots | Select-Object -Unique)) {
    if (-not (Test-Path -LiteralPath $sdkRoot -PathType Container)) { continue }
    $sdkRoot = (Get-Item -LiteralPath $sdkRoot).FullName
    foreach ($candidate in @(Get-VersionDirectories (Join-Path $sdkRoot 'Include') $SdkVersion)) {
        $version = $candidate.Name
        $required = @("Include\$version\ucrt\stdlib.h", "Include\$version\shared\winerror.h", "Include\$version\um\windows.h", "Include\$version\um\amsi.h", "Include\$version\um\aclapi.h", "Lib\$version\ucrt\$target\ucrt.lib")
        foreach ($library in @('kernel32.lib', 'amsi.lib', 'ole32.lib', 'advapi32.lib', 'uuid.lib')) { $required += "Lib\$version\um\$target\$library" }
        if ($BuildAcceptanceProbe) { $required += "Lib\$version\um\$target\userenv.lib" }
        if (Test-RequiredFiles $sdkRoot $required) { $sdk = $sdkRoot; $sdkVersionFound = $version; break }
    }
    if ($sdk) { break }
}
if (-not $sdk) { throw "A complete Windows SDK with AMSI headers and $target libraries was not found. Install the Windows SDK, or specify -WindowsSdkPath and -SdkVersion." }

$compilerDirectory = Join-Path $msvc "bin\Hostx64\$target"
$compiler = Join-Path $compilerDirectory 'cl.exe'
$includes = @((Join-Path $msvc 'include'), (Join-Path $sdk "Include\$sdkVersionFound\ucrt"), (Join-Path $sdk "Include\$sdkVersionFound\shared"), (Join-Path $sdk "Include\$sdkVersionFound\um"))
$winrt = Join-Path $sdk "Include\$sdkVersionFound\winrt"
if (Test-Path -LiteralPath $winrt -PathType Container) { $includes += $winrt }
$libraries = @((Join-Path $msvc "lib\$target"), (Join-Path $sdk "Lib\$sdkVersionFound\ucrt\$target"), (Join-Path $sdk "Lib\$sdkVersionFound\um\$target"))
$previousInclude = $env:INCLUDE
$previousLib = $env:LIB
$previousPath = $env:PATH
try {
    $env:INCLUDE = $includes -join ';'
    $env:LIB = $libraries -join ';'
    $env:PATH = "$compilerDirectory;$previousPath"
    Push-Location -LiteralPath $PSScriptRoot
    try {
        Write-Host "Building $Arch (MSVC target $target), tools $msvc, SDK $sdkVersionFound"
        $machine = $target.ToUpperInvariant()
        $compileArguments = @('/nologo', '/LD', '/EHsc', '/std:c++17', '/utf-8', '/O2', '/MT', '/DUNICODE', '/D_UNICODE', '/GS', 'utmstack_amsi.cpp', '/link', '/DEF:utmstack_amsi.def', "/MACHINE:$machine", 'amsi.lib', 'ole32.lib', 'advapi32.lib', 'uuid.lib', '/OUT:utmstack_amsi.dll')
        & $compiler @compileArguments
        $compilerExitCode = $LASTEXITCODE
        if ($compilerExitCode -ne 0) { throw "AMSI compiler/linker failed (exit $compilerExitCode)." }
        if (-not (Test-Path -LiteralPath 'utmstack_amsi.dll' -PathType Leaf)) { throw 'The compiler returned success without producing utmstack_amsi.dll.' }
        if ($BuildAcceptanceProbe) {
            & $compiler /nologo /EHsc /std:c++17 /utf-8 /O2 /MT /DUNICODE /D_UNICODE /GS acceptance_probe.cpp /link "/MACHINE:$machine" amsi.lib ole32.lib advapi32.lib uuid.lib userenv.lib /OUT:acceptance_probe.exe
            if ($LASTEXITCODE -ne 0) { throw "AMSI acceptance probe compiler/linker failed (exit $LASTEXITCODE)." }
            & $compiler /nologo /EHsc /std:c++17 /utf-8 /O2 /MT /GS bounded_stream_test.cpp /link "/MACHINE:$machine" /OUT:bounded_stream_test.exe
            if ($LASTEXITCODE -ne 0) { throw "AMSI stream test compiler/linker failed (exit $LASTEXITCODE)." }
        }
    } finally {
        Pop-Location
    }
} finally {
    $env:INCLUDE = $previousInclude
    $env:LIB = $previousLib
    $env:PATH = $previousPath
}
Write-Host "Built utmstack_amsi.dll for $Arch ($target). Signing and runtime acceptance are separate steps."
