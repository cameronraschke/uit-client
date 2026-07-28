if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
	$CommandLine = "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`""
	Start-Process powershell.exe -ArgumentList $CommandLine -Verb RunAs
	Exit
}

# TPM must be initialized before doing secure boot updates
# Do not reboot until TPM is ready
if (-not (Get-Tpm).TpmReady) {
	Write-Warning "TPM is not ready yet. Wait for auto-provisioning to complete."
	exit
}

# Set AvailableUpdates to 0x5944 (Triggers update)
Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot" -Name "AvailableUpdates" -Value 0x5944 -Type DWord

# Do Secure Boot (2023 Microsoft CA) update
Start-ScheduledTask -TaskName "\Microsoft\Windows\PI\Secure-Boot-Update"