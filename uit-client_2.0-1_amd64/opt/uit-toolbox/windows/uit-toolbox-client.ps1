Set-Variable -Name "computerInfoObj" -Value (Get-ComputerInfo)
Set-Variable -Name "win32CompSysObj" -Value (Get-CimInstance -Class Win32_ComputerSystem)
Set-Variable -Name "win32ComputerSystemProductObj" -Value (Get-CimInstance -Class Win32_ComputerSystemProduct)
# Set-Variable -Name "win32BiosObj" -Value (Get-CimInstance -Class Win32_BIOS)
# Set-Variable -Name "win32OperatingSystemObj" -Value (Get-CimInstance -Class Win32_OperatingSystem)
Set-Variable -Name "win32MemoryObj" -Value (Get-CimInstance -Class Win32_PhysicalMemory)
Set-Variable -Name "win32ProcessorObj" -Value (Get-CimInstance -Class Win32_Processor)
Set-Variable -Name "win32DiskDriveObj" -Value (Get-CimInstance -Class Win32_DiskDrive)
Set-Variable -Name "win32LogicalDiskObj" -Value (Get-CimInstance -Class Win32_LogicalDisk -Filter "DriveType = '3' AND Name = 'C:'")

$arr = @{
'tagnumber' = $null
'system_serial' = $win32ComputerSystemProductObj.IdentifyingNumber
# 'system_serial' = $computerInfoObj.BiosSeralNumber
'system_manufacturer' = $computerInfoObj.CsManufacturer
'system_model' = $computerInfoObj.CsModel
'system_sku' = $win32CompSysObj.SystemSKUNumber
'chassis_type' = $computerInfoObj.CsPCSystemType
'bios_version' = $computerInfoObj.BiosSMBIOSBIOSVersion
'bios_release_date' = (Get-Date -Date $computerInfoObj.BiosReleaseDate).ToString("yyyy-MM-dd'T'HH:mm:ssK")
'tpm_version' = (Get-WmiObject -Namespace "Root\CIMv2\Security\MicrosoftTpm" -Class Win32_Tpm | Select-Object -ExpandProperty SpecVersion) -split ", " | Select-Object -First 1
'os_installed_at' = (Get-Date -Date ($computerInfoObj.OsInstallDate)).ToString("yyyy-MM-dd'T'HH:mm:ssK")
'os_vendor' = $computerInfoObj.OSManufacturer
'os_platform' = $computerInfoObj.OsType
'os_architecture' = $computerInfoObj.OsArchitecture
'os_name' = $computerInfoObj.OSName
'os_version' = $computerInfoObj.OSVersion
'windows_display_version' = $computerInfoObj.OSDisplayVersion
'windows_build_number' = $computerInfoObj.OsBuildNumber
'windows_ubr' = (Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name UBR).UBR
'windows_bitlocker_enabled' = (Get-BitLockerVolume -MountPoint "C:").VolumeStatus -eq "FullyEncrypted"
'ad_domain' = $computerInfoObj.Domain
'ad_domain_user' = $win32CompSysObj.DNSHostName
'memory_capacity_kb' = (Get-CimInstance -Class Win32_PhysicalMemory | Measure-Object -Property Capacity -Sum).Sum / 1024
'memory_speed_mhz' = $win32MemoryObj.Speed | Select-Object -First 1
'cpu_model' = $win32ProcessorObj.Name
'cpu_core_count' = $win32ProcessorObj.NumberOfCores
'cpu_thread_count' = $win32ProcessorObj.NumberOfLogicalProcessors
'disk_model' = ($win32DiskDriveObj | Select-Object -ExpandProperty Model | Select-Object -First 1)
'disk_type' = (Get-Disk | Where-Object { ($_.DiskNumber -eq "0") } | Select-Object -ExpandProperty BusType).ToLower()
'disk_size_kb' = ($win32DiskDriveObj | Measure-Object -Property Size -Sum).Sum / 1024
'disk_free_space_kb' = ($win32LogicalDiskObj | Measure-Object -Property FreeSpace -Sum).Sum / 1024
'ethernet_mac_addr' = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
'wifi_mac_addr' = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled -and $_.Description -match "Wireless" } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
# 'cpuTemp' = (Get-CimInstance -Namespace root\wmi -Class MSAcpi_ThermalZoneTemperature | Select-Object -ExpandProperty CurrentTemperature | Select-Object -First 1) / 10 - 273.15
'battery_manufacturer' = (Get-WmiObject -Namespace "root\wmi" -Class "BatteryStaticData").ManufactureName
'battery_serial' = (Get-CimInstance -Class Win32_Battery).SerialNumber
# 'battery_charge_percent' = (Get-CimInstance -Class Win32_Battery).EstimatedChargeRemaining
'battery_current_max_capacity' = (Get-CimInstance -Namespace "root\wmi" -ClassName "BatteryFullChargedCapacity").BatteryFullChargedCapacity
'battery_design_capacity' = (Get-WmiObject -Namespace "root\wmi" -Class "BatteryStaticData").DesignedCapacity
'battery_health' = (((Get-CimInstance -Namespace "root\wmi" -Class "BatteryFullChargedCapacity").BatteryFullChargedCapacity) / (Get-WmiObject -Namespace "root\wmi" -Class "BatteryStaticData").DesignedCapacity) * 100
'battery_charge_cycles' = (Get-WmiObject -Namespace "root\wmi" -ClassName BatteryCycleCount).CycleCount
'updated_from_windows' = $true
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
$okayToGo = Read-Host "You entered tag number $tagNum. Is this correct? (Y/N)"
if ($okayToGo -ne "Y") {
	Write-Host "Exiting. Please run the script again and enter the correct tag number."
	exit
}

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