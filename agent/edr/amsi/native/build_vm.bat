@echo off
for /d %%v in ("C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Tools\MSVC\*") do set MSVCDIR=%%v
set SDK=C:\Program Files (x86)\Windows Kits\10
set SDKVER=10.0.22621.0
set INCLUDE=%MSVCDIR%\include;%SDK%\Include\%SDKVER%\ucrt;%SDK%\Include\%SDKVER%\shared;%SDK%\Include\%SDKVER%\um;%SDK%\Include\%SDKVER%\winrt
set LIB=%MSVCDIR%\lib\arm64;%SDK%\Lib\%SDKVER%\ucrt\arm64;%SDK%\Lib\%SDKVER%\um\arm64
cd /d C:\edrbuild
"%MSVCDIR%\bin\Hostx64\arm64\cl.exe" /nologo /LD /EHsc /O2 /DUNICODE /D_UNICODE /GS utmstack_amsi.cpp /link /DEF:utmstack_amsi.def amsi.lib ole32.lib /OUT:utmstack_amsi.dll
