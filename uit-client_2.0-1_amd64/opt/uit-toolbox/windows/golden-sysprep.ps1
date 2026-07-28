if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
	$CommandLine = "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`""
	Start-Process powershell.exe -ArgumentList $CommandLine -Verb RunAs
	Exit
}

# Turn off hibernation
powercfg /h off

# Disable bitlocker
manage-bde -off C:

# Add AD FOD
Add-WindowsCapability -Online -Name "Rsat.ActiveDirectory.DS-LDS.Tools~~~~0.0.1.0"

# Shrink image size
C:\WINDOWS\system32\DISM.exe /Online /Cleanup-Image /StartComponentCleanup /ResetBase

# Clear IDs
Remove-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" -Name "SUSClientId" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" -Name "SusClientIdValidation" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" -Name "AccountDomainSid" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate" -Name "PingID" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Cryptography" -Name "MachineGuid" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\SQMClient" -Name "MachineId" -ErrorAction SilentlyContinue
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Diagnostics\DiagTrack\*" -Recurse -Force -ErrorAction SilentlyContinue

# Secure boot
Remove-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot" -Name "AvailableUpdates" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot" -Name "MicrosoftUpdateManagedOptIn" -ErrorAction SilentlyContinue
Remove-Item -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot\Servicing" -Recurse -ErrorAction SilentlyContinue

# Clear group policy
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Group Policy\History\*" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Group Policy\State\*" -Recurse -Force -ErrorAction SilentlyContinue

# Clear network adapters history
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Profiles\*" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Signatures\Unmanaged\*" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\NetworkList\Signatures\Managed\*" -Recurse -Force -ErrorAction SilentlyContinue

# Clear mounted disks history
Get-Item -Path "HKLM:\SYSTEM\MountedDevices" | Clear-ItemProperty -ErrorAction SilentlyContinue

# Clear pending file rename operations
Remove-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager" -Name "PendingFileRenameOperations" -ErrorAction SilentlyContinue

Write-Host "Master image registry cleanup complete!" -ForegroundColor Green

Write-Host "Sysprep will now run. This will take a few minutes. Please wait..." -ForegroundColor Green

C:\WINDOWS\system32\Sysprep\Sysprep.exe /generalize /oobe /shutdown /unattend:'C:\WINDOWS\system32\Sysprep\unattend.xml'