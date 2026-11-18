# Produce a hash-pinned service/provider pair. This does not sign or install it.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$ServiceBinary,
    [Parameter(Mandatory)][ValidateSet('amd64','arm64')][string]$Arch,
    [Parameter(Mandatory)][string]$PackageVersion,
    [Parameter(Mandatory)][string]$Implementation,
    [Parameter(Mandatory)][string]$OutputDirectory,
    [string]$ProviderDll=(Join-Path $PSScriptRoot 'utmstack_amsi.dll'),
    [ValidateSet('LabOnly','Production')][string]$Scope='LabOnly',
    [string]$ExpectedSignerThumbprint
)
Set-StrictMode -Version Latest
$ErrorActionPreference='Stop'
function Get-PeMachine([string]$Path){
    $stream=[IO.File]::Open($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
    $reader=New-Object IO.BinaryReader($stream)
    try{
        if($stream.Length -lt 64 -or $reader.ReadUInt16() -ne 0x5A4D){throw "Not a PE file: $Path"}
        $stream.Position=0x3C;$offset=$reader.ReadInt32()
        if($offset -lt 64 -or [long]$offset+6 -gt $stream.Length){throw "Invalid PE header offset: $Path"}
        $stream.Position=$offset
        if($reader.ReadUInt32() -ne 0x00004550){throw "Invalid PE signature: $Path"}
        return $reader.ReadUInt16()
    }finally{$reader.Dispose();$stream.Dispose()}
}
$expectedMachine=(@{amd64=0x8664;arm64=0xAA64})[$Arch]
# Each architecture is packaged in its own output directory. The provider
# keeps the runtime filename used by service/service.go and releasepack.
$files=@([pscustomobject]@{source=$ServiceBinary;name="utmstack_edr_windows_$Arch.exe"},[pscustomobject]@{source=$ProviderDll;name='utmstack_amsi.dll'})
if($Scope -eq 'Production' -and -not $ExpectedSignerThumbprint){throw 'Production packaging requires an explicitly pinned Authenticode signer thumbprint.'}
New-Item -ItemType Directory -Force -Path $OutputDirectory|Out-Null
$outputItem=Get-Item -LiteralPath $OutputDirectory
if(-not $outputItem.PSIsContainer -or ($outputItem.Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Package output must be a regular directory, not a reparse point.'}
$manifestPath=Join-Path $OutputDirectory 'pair-manifest.json'
# A failed rebuild must not leave an older manifest describing a partial pair.
if(Test-Path -LiteralPath $manifestPath){
    $oldManifest=Get-Item -LiteralPath $manifestPath
    if($oldManifest.PSIsContainer -or ($oldManifest.Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Package manifest must be a regular file.'}
    Remove-Item -LiteralPath $manifestPath -Force
}
$artifacts=@()
foreach($file in $files){
    $item=Get-Item -LiteralPath $file.source
    if($item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Package inputs must be regular files, not reparse points.'}
    $destination=Join-Path $OutputDirectory $file.name
    if([IO.Path]::GetFullPath($item.FullName) -eq [IO.Path]::GetFullPath($destination)){throw 'Output must differ from the source input.'}
    if(Test-Path -LiteralPath $destination){
        $oldOutput=Get-Item -LiteralPath $destination
        if($oldOutput.PSIsContainer -or ($oldOutput.Attributes -band [IO.FileAttributes]::ReparsePoint)){throw 'Package artifacts must be regular files, not reparse points.'}
    }
    # Hold the source against writes/deletion while Copy-Item reads it. Validate
    # the copied bytes, never attach an earlier source signature to new bytes.
    $sourceLock=[IO.File]::Open($item.FullName,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
    try{Copy-Item -LiteralPath $item.FullName -Destination $destination -Force}finally{$sourceLock.Dispose()}
    $outputLock=[IO.File]::Open($destination,[IO.FileMode]::Open,[IO.FileAccess]::Read,[IO.FileShare]::Read)
    try{
        if((Get-PeMachine $destination) -ne $expectedMachine){throw "Wrong PE machine for $Arch`: $($item.Name)"}
        $signature=Get-AuthenticodeSignature -LiteralPath $destination
        if($Scope -eq 'Production' -and ($signature.Status -ne 'Valid' -or $signature.SignerCertificate.Thumbprint -ne $ExpectedSignerThumbprint)){throw "Production input does not have the pinned valid signer: $($item.Name)"}
        $artifacts += [ordered]@{file=$file.name;bytes=$outputLock.Length;sha256=(Get-FileHash -Algorithm SHA256 $destination).Hash.ToLowerInvariant();authenticode_status=[string]$signature.Status;signer_thumbprint=$(if($signature.SignerCertificate){$signature.SignerCertificate.Thumbprint}else{$null})}
    }finally{$outputLock.Dispose()}
}
$manifest=[ordered]@{schema=1;scope=$Scope;package_version=$PackageVersion;implementation=$Implementation;os='windows';architecture=$Arch;amsi_wire_protocol='UTA2';deployment_requirement='Install the service executable and provider DLL together; keep provider filename utmstack_amsi.dll beside the executable at runtime.';artifacts=$artifacts}|ConvertTo-Json -Depth 8
[IO.File]::WriteAllText($manifestPath,$manifest,(New-Object Text.UTF8Encoding($false)))
Write-Host "Prepared $Scope $Arch pair. Deployment and production signing acceptance are separate."
