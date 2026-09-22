# Test script for rule:
#   "Windows: Multiple Logon Failure Followed by Logon Success"
#   (bruteforce_multiple_logon_failure_followed_by_success.yml)
#
# Rule trigger conditions:
#   - 10+ failed logons (EventID 4625) followed by a successful logon (EventID 4624)
#   - Same target.user / origin.host / origin.ip
#   - Within a 5 minute window
#
# Requirements:
#   - Run as Administrator on a Windows host shipping the Security log to UTMStack.
#   - Audit policy must log both success and failure for "Logon":
#       auditpol /set /subcategory:"Logon" /success:enable /failure:enable
#
# Usage:
#   powershell -ExecutionPolicy Bypass -File .\bruteforce_multiple_logon_failure_followed_by_success.test.ps1

#Requires -RunAsAdministrator
$ErrorActionPreference = 'Stop'

$TestUser      = 'utm_bf_test'
$CorrectPass   = 'C0rrect-P@ssw0rd!_' + (Get-Random -Maximum 99999)
$WrongPass     = 'WrongP@ssword_xyz'
$FailureCount  = 12   # rule needs 10; do a couple extra for safety
$DelayMs       = 400

Write-Host "[*] Ensuring local test user '$TestUser' exists..."
$securePass = ConvertTo-SecureString $CorrectPass -AsPlainText -Force
$existing   = Get-LocalUser -Name $TestUser -ErrorAction SilentlyContinue
if ($null -eq $existing) {
    New-LocalUser -Name $TestUser -Password $securePass `
                  -PasswordNeverExpires -AccountNeverExpires `
                  -Description 'UTMStack bruteforce rule test (safe to delete)' | Out-Null
    Add-LocalGroupMember -Group 'Users' -Member $TestUser
} else {
    Set-LocalUser -Name $TestUser -Password $securePass
}

Write-Host "[*] Loading native LogonUser API..."
Add-Type -Namespace UTM -Name Native -MemberDefinition @"
[System.Runtime.InteropServices.DllImport("advapi32.dll", SetLastError=true, CharSet=System.Runtime.InteropServices.CharSet.Unicode)]
public static extern bool LogonUser(
    string lpszUsername,
    string lpszDomain,
    string lpszPassword,
    int    dwLogonType,
    int    dwLogonProvider,
    out System.IntPtr phToken);

[System.Runtime.InteropServices.DllImport("kernel32.dll", SetLastError=true)]
public static extern bool CloseHandle(System.IntPtr hObject);
"@

$LOGON32_LOGON_NETWORK    = 3   # generates clean 4624/4625 with workstation/IP fields populated
$LOGON32_PROVIDER_DEFAULT = 0
$domain                   = $env:COMPUTERNAME

function Invoke-Logon {
    param([string]$User, [string]$Pass, [string]$Tag)
    $token = [IntPtr]::Zero
    $ok = [UTM.Native]::LogonUser($User, $domain, $Pass,
                                  $LOGON32_LOGON_NETWORK,
                                  $LOGON32_PROVIDER_DEFAULT,
                                  [ref]$token)
    $code = [System.Runtime.InteropServices.Marshal]::GetLastWin32Error()
    Write-Host ("    {0,-8} success={1} lastErr={2}" -f $Tag, $ok, $code)
    if ($ok -and $token -ne [IntPtr]::Zero) {
        [UTM.Native]::CloseHandle($token) | Out-Null
    }
    return $ok
}

Write-Host "[*] Generating $FailureCount failed logons (EventID 4625)..."
for ($i = 1; $i -le $FailureCount; $i++) {
    Invoke-Logon -User $TestUser -Pass $WrongPass -Tag ("fail#$i") | Out-Null
    Start-Sleep -Milliseconds $DelayMs
}

Write-Host "[*] Generating 1 successful logon (EventID 4624)..."
$success = Invoke-Logon -User $TestUser -Pass $CorrectPass -Tag 'success'
if (-not $success) {
    Write-Warning "Final logon failed. Check audit policy and account state."
}

Write-Host ""
Write-Host "[*] Recent 4625/4624 events for $TestUser :"
Get-WinEvent -FilterHashtable @{
    LogName   = 'Security'
    Id        = 4624, 4625
    StartTime = (Get-Date).AddMinutes(-5)
} -ErrorAction SilentlyContinue |
    Where-Object { $_.Message -match [regex]::Escape($TestUser) } |
    Select-Object TimeCreated, Id, @{n='User';e={$TestUser}} |
    Format-Table -AutoSize

Write-Host ""
Write-Host "[+] Done. UTMStack should fire:"
Write-Host "    'Windows: Multiple Logon Failure Followed by Logon Success'"
Write-Host ""
Write-Host "[i] To remove the test user when finished:"
Write-Host "    Remove-LocalUser -Name $TestUser"
