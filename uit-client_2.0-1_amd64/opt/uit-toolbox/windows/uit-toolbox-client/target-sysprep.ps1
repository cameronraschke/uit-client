if (-not ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
	$CommandLine = "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`""
	Start-Process powershell.exe -ArgumentList $CommandLine -Verb RunAs
	Exit
}

# TPM must be initialized before doing secure boot updates
# Do not reboot until TPM is ready
if (-not (Get-Tpm).TpmReady) {
	Write-Warning "TPM is not ready yet. Wait for auto-provisioning to complete."
	Read-Host "Press Enter to continue after TPM is ready..."
	Exit
}

$secureBootUpdatedInRegistry = $false
$newKeyInSecureBootDB = $false
$secureBootUpdatedInRegistry = (Get-ItemPropertyValue -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot\Servicing" -Name "UEFICA2023Status"  -ErrorAction SilentlyContinue) -eq "Updated"
$newKeyInSecureBootDB = [System.Text.Encoding]::ASCII.GetString((Get-SecureBootUEFI db).bytes) -match 'Windows UEFI CA 2023'
if (-not ($secureBootUpdatedInRegistry -and $newKeyInSecureBootDB)) {
	# Set AvailableUpdates to 0x5944 (Triggers update)
	$currentUpdateValue = (Get-ItemPropertyValue -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot" -Name "AvailableUpdates" -ErrorAction SilentlyContinue).toString("X")
	if ($currentUpdateValue -ne "5944") {
		Write-Host "Setting AvailableUpdates to 0x5944 to trigger Secure Boot update..."
		Set-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot" -Name "AvailableUpdates" -Value 0x5944 -Type DWord
	}
	
	# Do Secure Boot update (2023 Microsoft CA)
	if (-not (Get-ScheduledTask -TaskName "Secure-Boot-Update" -ErrorAction SilentlyContinue).State -eq "Running") {
		Write-Host "Starting Secure Boot update task..."
		Start-ScheduledTask -TaskName "\Microsoft\Windows\PI\Secure-Boot-Update"
	} else {
		Write-Host "Secure Boot update task is already running."
	}
}

Read-Host "Post-Sysprep scripts finished, press Enter to continue..."
