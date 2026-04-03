$arr = @{
'system_serial' = (Get-CimInstance -Class Win32_ComputerSystemProduct).IdentifyingNumber
'chassis_type' = (Get-CimInstance -Class Win32_ComputerSystem).ChassisSKUNumber
'ad_domain' = (Get-CimInstance -Class Win32_ComputerSystem).Domain
'ad_domain_joined' = (($null -ne $adDomain) -and ($adDomain -ne ""))
'system_manufacturer' = (Get-CimInstance -Class Win32_ComputerSystem).Manufacturer
'system_model' = (Get-CimInstance -Class Win32_ComputerSystem).Model
'bios_version' = (Get-CimInstance -Class Win32_BIOS).SMBIOSBIOSVersion
'os_name' = (Get-CimInstance -Class Win32_OperatingSystem).Caption
'os_installed_at' = (Get-Date -Date ((Get-CimInstance -Class Win32_OperatingSystem).InstallDate) -UFormat '%Y-%m-%dT%H:%M:%S%Z')
'os_version' = (Get-CimInstance -Class Win32_OperatingSystem).BuildNumber
'os_ubr' = (Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name UBR).UBR
'memory_capacity_kb' = (Get-CimInstance -Class Win32_PhysicalMemory | Measure-Object -Property Capacity -Sum).Sum / 1024
'memory_speed_mhz' = (Get-CimInstance -Class Win32_PhysicalMemory).Speed | Select-Object -First 1
'cpu_model' = (Get-CimInstance -Class Win32_Processor).Name
'cpu_cores' = (Get-CimInstance -Class Win32_Processor).NumberOfCores
'cpu_threads' = (Get-CimInstance -Class Win32_Processor).NumberOfLogicalProcessors
'disk_size_kb' = (Get-CimInstance -Class Win32_DiskDrive | Measure-Object -Property Size -Sum).Sum / 1024
'ethernet_mac_addr' = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
'wifi_mac_addr' = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled -and $_.Description -match "Wireless" } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
'disk_model' = (Get-CimInstance -Class Win32_DiskDrive | Select-Object -ExpandProperty Model | Select-Object -First 1)
# 'cpuTemp' = (Get-CimInstance -Namespace root\wmi -Class MSAcpi_ThermalZoneTemperature | Select-Object -ExpandProperty CurrentTemperature | Select-Object -First 1) / 10 - 273.15
'battery_charge_percent' = (Get-CimInstance -Class Win32_Battery).EstimatedChargeRemaining
}

$httpBodyArr = @{}
foreach ($key in $arr.Keys) {
	if ($null -ne $arr[$key]) {
		$httpBodyArr[$key] = $arr[$key]
	}
	if ($null -eq $arr[$key]) {
		$httpBodyArr[$key] = $null
	}
	if ($arr[$key] -is [string]) {
		$httpBodyArr[$key] = $arr[$key].Trim()
	}
	if ($arr[$key] -is [string] -and [string]::IsNullOrWhiteSpace($arr[$key])) {
		$httpBodyArr[$key] = $null
	}
	if ($key -eq "ethernet_mac_addr" -and $null -ne $httpBodyArr[$key]) {
		$httpBodyArr[$key] = $httpBodyArr[$key].Replace("-", ":")
	}
	if ($key -eq "wifi_mac_addr" -and $null -ne $httpBodyArr[$key]) {
		$httpBodyArr[$key] = $httpBodyArr[$key].Replace("-", ":")
	}
}

$tagNum = Read-Host "Enter tag number (100000-999999)"

$httpBodyArr["tagnumber"] = $tagNum

Add-Type -AssemblyName System.Windows.Forms
$dialogObj = New-Object System.Windows.Forms.FolderBrowserDialog
$dialogObj.Description = "Select a folder to save the backups"
$fileDialog = $dialogObj.ShowDialog()

if ($fileDialog -eq [System.Windows.Forms.DialogResult]::OK -and 
	-not [string]::IsNullOrWhiteSpace($dialogObj.SelectedPath)) {
	Write-Host "Selected folder: $($dialogObj.SelectedPath)"
} else {
	Write-Host "No folder selected. Exiting."
	exit
}

$jsonStr = $httpBodyArr | ConvertTo-Json -Depth 4
Out-File -FilePath "$($dialogObj.SelectedPath)\uit-system-info.json" -InputObject $jsonStr -Encoding UTF8
Write-Host $jsonStr