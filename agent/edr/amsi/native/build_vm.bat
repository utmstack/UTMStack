@echo off
setlocal
rem Use the same discovery, target mapping and exit-code checks in a VM.
rem Default target remains arm64; pass -Arch amd64 for the x64 build.
powershell.exe -NoLogo -NoProfile -NonInteractive -File "%~dp0build.ps1" %*
exit /b %ERRORLEVEL%
