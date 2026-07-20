@echo off

:: C:\Windows\Setup\Scripts\SetupComplete.cmd

:: Turn hibernation back on
powercfg /h on

:: TPM reset, re-initializes TPM on next boot
powershell -Command "Clear-Tpm"

:: Clear windows update identifiers
net stop wuauserv
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" /v SusClientId /f
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" /v SusClientIdValidation /f
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" /v PingID /f
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" /v AccountDomainSid /f
net start wuauserv

:: Clear machine GUIDs
reg delete "HKLM\SOFTWARE\Microsoft\Cryptography" /v MachineGuid /f
reg delete "HKLM\SOFTWARE\Microsoft\SQMClient" /v MachineId /f

:: Clear GPOs 
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Group Policy\History" /f
reg delete "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Group Policy\State" /f

:: Prepare 2023 secure boot CA
reg add "HKLM\SYSTEM\CurrentControlSet\Control\SecureBoot" /v "AvailableUpdates" /t REG_DWORD /d 0x5944 /f
schtasks /run /tn "\Microsoft\Windows\PI\Secure-Boot-Update"