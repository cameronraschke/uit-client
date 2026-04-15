Set-Variable -Name "computerInfoObj" -Value (Get-ComputerInfo)
Set-Variable -Name "win32CompSysObj" -Value (Get-CimInstance -Class Win32_ComputerSystem)
Set-Variable -Name "win32ComputerSystemProductObj" -Value (Get-CimInstance -Class Win32_ComputerSystemProduct)
# Set-Variable -Name "win32BiosObj" -Value (Get-CimInstance -Class Win32_BIOS)
# Set-Variable -Name "win32OperatingSystemObj" -Value (Get-CimInstance -Class Win32_OperatingSystem)
Set-Variable -Name "win32MemoryObj" -Value (Get-CimInstance -Class Win32_PhysicalMemory)
Set-Variable -Name "win32ProcessorObj" -Value (Get-CimInstance -Class Win32_Processor)
Set-Variable -Name "win32DiskDriveObj" -Value (Get-CimInstance -Class Win32_DiskDrive -Filter "MediaType != 'Removable Media' AND interfaceType = 'SCSI'")
Set-Variable -Name "win32LogicalDiskObj" -Value (Get-CimInstance -Class Win32_LogicalDisk -Filter "DriveType = '3' AND Name = 'C:'")
Set-Variable -Name "win32BatteryObj" -Value (Get-CimInstance -Class Win32_Battery -ErrorAction SilentlyContinue)
Set-Variable -Name "batteryStaticDataObj" -Value (Get-WmiObject -Namespace "root\wmi" -Class "BatteryStaticData" -ErrorAction SilentlyContinue)
Set-Variable -Name "batteryCycleCountObj" -Value (Get-WmiObject -Namespace "root\wmi" -ClassName BatteryCycleCount -ErrorAction SilentlyContinue)
Set-Variable -Name "batteryFullChargedCapacityObj" -Value (Get-CimInstance -Namespace "root\wmi" -ClassName "BatteryFullChargedCapacity" -ErrorAction SilentlyContinue)

$arr = @{}

# Tag number
$arr['tagnumber'] = $null
$tagNum = Read-Host "Enter tag number (100000-999999)"
$okayToGo = Read-Host "You entered tag number $tagNum. Is this correct? (Y/N)"
if ($okayToGo -ne "Y") {
	Write-Host "Exiting. Please run the script again and enter the correct tag number."
	exit
}
if ($tagNum -notmatch "^\d{6}$" -or [int]$tagNum -lt 100000 -or [int]$tagNum -gt 999999) {
	Write-Host "Invalid tag number. Please enter a 6-digit number between 100000 and 999999."
	exit
}
$arr["tagnumber"] = [System.Int64]$tagNum

# System serial
$arr['system_serial'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32ComputerSystemProductObj.IdentifyingNumber)) {
	$arr['system_serial'] = [System.String]$win32ComputerSystemProductObj.IdentifyingNumber.Trim()
} elseif (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.BiosSeralNumber)) {
	# This is misspelled in PowerShell
	$arr['system_serial'] = [System.String]$computerInfoObj.BiosSeralNumber.Trim()
} else {
	Write-Host "System serial number not found in WMI."
}

# System manufacturer
$arr['system_manufacturer'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsManufacturer)) {
	$arr['system_manufacturer'] = [System.String]$computerInfoObj.CsManufacturer.Trim()
} else {
	Write-Host "System manufacturer not found in WMI."
}

# System model
$arr['system_model'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsModel)) {
	$arr['system_model'] = [System.String]$computerInfoObj.CsModel.Trim()
} else {
	Write-Host "System model not found in WMI."
}

# System SKU
$arr['system_sku'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32CompSysObj.SystemSKUNumber)) {
	$arr['system_sku'] = [System.String]$win32CompSysObj.SystemSKUNumber.Trim()
} else {
	Write-Host "System SKU not found in WMI."
}

# Chassis type
$arr['chassis_type'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsPCSystemType)) {
	$arr['chassis_type'] = ([System.String]$computerInfoObj.CsPCSystemType).Trim()
} else {
	Write-Host "Chassis type not found in WMI."
}

# BIOS version
$arr['bios_version'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.BiosSMBIOSBIOSVersion)) {
	$arr['bios_version'] = [System.String]$computerInfoObj.BiosSMBIOSBIOSVersion.Trim()
} else {
	Write-Host "BIOS version not found in WMI."
}

# BIOS release date
$arr['bios_release_date'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.BiosReleaseDate)) {
	$parsedBiosDate = [System.DateTime]::MinValue
	if ([System.DateTime]::TryParse($computerInfoObj.BiosReleaseDate, [ref]$parsedBiosDate)) {
		$arr['bios_release_date'] = [System.String]$parsedBiosDate.ToString("yyyy-MM-dd'T'HH:mm:sszzzz")
	} else {
	Write-Host "Failed to parse BIOS release date."
		$arr['bios_release_date'] = $null
	}
} else {
	Write-Host "BIOS release date not found in WMI."
}

#TPM version
$arr['tpm_version'] = $null
try {
	$tpmVersion = (Get-WmiObject -Namespace "Root\CIMv2\Security\MicrosoftTpm" -Class Win32_Tpm | Select-Object -ExpandProperty SpecVersion) -split ", " | Select-Object -First 1
	if (-not [System.String]::IsNullOrWhiteSpace($tpmVersion)) {
		$arr['tpm_version'] = [System.String]$tpmVersion
	} else {
		Write-Host "TPM version not found in WMI."
	}
} catch {
	Write-Host "Error retrieving TPM version: $_"
}

# OS Install date
$arr['os_installed_at'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OsInstallDate)) {
	$parsedOSInstallDate = [System.DateTime]::MinValue
	if ([System.DateTime]::TryParse($computerInfoObj.OsInstallDate, [ref]$parsedOSInstallDate)) {
		$arr['os_installed_at'] = [System.String]$parsedOSInstallDate.ToString("yyyy-MM-dd'T'HH:mm:sszzzz")
	} else {
		Write-Host "Failed to parse OS install date."
	}
} else {
	Write-Host "OS install date not found in WMI."
}

# OS vendor
$arr['os_vendor'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSManufacturer)) {
	$arr['os_vendor'] = [System.String]$computerInfoObj.OSManufacturer.Trim()
} else {
	Write-Host "OS vendor not found in WMI."
}

# OS platform
$arr['os_platform'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OsType)) {
	$arr['os_platform'] = ([System.String]$computerInfoObj.OsType).Trim()
} else {
	Write-Host "OS platform not found in WMI."
}

# OS architecture
$arr['os_architecture'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OsArchitecture)) {
	$arr['os_architecture'] = [System.String]$computerInfoObj.OsArchitecture.Trim()
} else {
	Write-Host "OS architecture not found in WMI."
}

# OS name
$arr['os_name'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSName)) {
	$arr['os_name'] = [System.String]$computerInfoObj.OSName.Trim()
} else {
	Write-Host "OS name not found in WMI."
}

# OS version
$arr['os_version'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSVersion)) {
	$arr['os_version'] = [System.String]$computerInfoObj.OSVersion.Trim()
} else {
	Write-Host "OS version not found in WMI."
}

# Windows display version
$arr['windows_display_version'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSDisplayVersion)) {
	$arr['windows_display_version'] = [System.String]$computerInfoObj.OSDisplayVersion.Trim()
} else {
	Write-Host "Windows display version not found in WMI."
}

# Windows build number
$arr['windows_build_number'] = $null
$windowsBuildNumberRaw = $computerInfoObj.OsBuildNumber
if (-not [System.String]::IsNullOrWhiteSpace($windowsBuildNumberRaw)) {
	$windowsBuildNumber = [System.Int64]0
	if ([System.Int64]::TryParse([string]$windowsBuildNumberRaw, [ref]$windowsBuildNumber) -and $windowsBuildNumber -gt 0) {
		$arr['windows_build_number'] = $windowsBuildNumber
	} else {
		Write-Host "Windows build number not found in WMI."
	}
} else {
	Write-Host "Windows build number not found in WMI."
}

# Windows UBR
$arr['windows_ubr'] = $null
try {
	$ubrValue = (Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name UBR).UBR
	if ([System.Int64]::TryParse($ubrValue, [ref]$null)) {
		$arr['windows_ubr'] = [System.Int64]$ubrValue
	} else {
		Write-Host "Windows UBR value not found in registry."
	}
} catch {
	Write-Host "Error retrieving Windows UBR value from registry: $_"
}

# Windows BitLocker enabled
$arr['windows_bitlocker_enabled'] = $null
try {
	$bitlockerStatus = (Get-BitLockerVolume -MountPoint "C:").VolumeStatus
	if (-not [System.String]::IsNullOrWhiteSpace($bitlockerStatus)) {
		$arr['windows_bitlocker_enabled'] = [System.Boolean]($bitlockerStatus -eq "FullyEncrypted")
	} else {
		Write-Host "BitLocker status not found."
	}
} catch {
	Write-Host "Error retrieving BitLocker status: $_"
}

# AD domain
$arr['ad_domain'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsDomain)) {
	$arr['ad_domain'] = [System.String]($computerInfoObj.CsDomain).Trim()
} else {
	Write-Host "AD domain not found in WMI"
}

# AD domain user
$arr['ad_domain_user'] = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsDNSHostName)) {
	$arr['ad_domain_user'] = [System.String]($computerInfoObj.CsDNSHostName).Trim()
} else {
	Write-Host "AD domain user not found in WMI."
}

# AD distinguished name
$arr['ad_distinguished_name'] = $null
Get-ADComputer -Identity $env:COMPUTERNAME -Properties DistinguishedName -ErrorAction SilentlyContinue | ForEach-Object {
	if (-not [System.String]::IsNullOrWhiteSpace($_.DistinguishedName)) {
		$arr['ad_distinguished_name'] = [System.String]$_.DistinguishedName.Trim()
	} else {
		Write-Host "AD distinguished name not found for computer $env:COMPUTERNAME."
	}
}

# Memory capacity in KB
$arr['memory_capacity_kb'] = $null
try {
	$memoryCapacityBytes = ($win32MemoryObj | Measure-Object -Property Capacity -Sum).Sum
	if ($null -ne $memoryCapacityBytes -and [System.Int64]$memoryCapacityBytes -gt 0) {
		$arr['memory_capacity_kb'] = [System.Int64]($memoryCapacityBytes / 1024)
	} else {
		Write-Host "Memory capacity not found."
	}
} catch {
	Write-Host "Error retrieving memory capacity: $_"
}

# memory speed in MHz
$arr['memory_speed_mhz'] = $null
try {
	$memorySpeedRaw = ($win32MemoryObj | Select-Object -ExpandProperty Speed | Select-Object -First 1)
	$memorySpeed = [System.Int64]0
	if ([System.Int64]::TryParse([string]$memorySpeedRaw, [ref]$memorySpeed) -and $memorySpeed -gt 0) {
		$arr['memory_speed_mhz'] = $memorySpeed
	} else {
		Write-Host "Memory speed not found."
	}
} catch {
	Write-Host "Error retrieving memory speed: $_"
}

# CPU model
$arr['cpu_model'] = $null
try {
	$cpuModel = ($win32ProcessorObj | Select-Object -ExpandProperty Name | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($cpuModel)) {
		$arr['cpu_model'] = [System.String]$cpuModel.Trim()
	} else {
		Write-Host "CPU model not found."
	}
} catch {
	Write-Host "Error retrieving CPU model: $_"
}

# CPU core count
$arr['cpu_core_count'] = $null
try {
	$cpuCoreCountRaw = ($win32ProcessorObj | Select-Object -ExpandProperty NumberOfCores | Select-Object -First 1)
	$cpuCoreCount = [System.Int64]0
	if ([System.Int64]::TryParse([string]$cpuCoreCountRaw, [ref]$cpuCoreCount) -and $cpuCoreCount -gt 0) {
		$arr['cpu_core_count'] = $cpuCoreCount
	} else {
		Write-Host "CPU core count not found."
	}
} catch {
	Write-Host "Error retrieving CPU core count: $_"
}

# CPU thread count
$arr['cpu_thread_count'] = $null
try {
	$cpuThreadCountRaw = ($win32ProcessorObj | Select-Object -ExpandProperty NumberOfLogicalProcessors | Select-Object -First 1)
	$cpuThreadCount = [System.Int64]0
	if ([System.Int64]::TryParse([string]$cpuThreadCountRaw, [ref]$cpuThreadCount) -and $cpuThreadCount -gt 0) {
		$arr['cpu_thread_count'] = $cpuThreadCount
	} else {
		Write-Host "CPU thread count not found."
	}
} catch {
	Write-Host "Error retrieving CPU thread count: $_"
}

# Disk model
$arr['disk_model'] = $null
try {
	$diskModel = ($win32DiskDriveObj | Select-Object -ExpandProperty Model | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($diskModel)) {
		$arr['disk_model'] = [System.String]$diskModel.Trim()
	} else {
		Write-Host "Disk model not found. Setting disk_model to null."
		$arr['disk_model'] = $null
	}
} catch {
	Write-Host "Error retrieving disk model: $_. Setting disk_model to null."
	$arr['disk_model'] = $null
}

# Disk type
$arr['disk_type'] = $null
try {
	$diskType = (Get-Disk | Where-Object { $_.DiskNumber -eq "0" } | Select-Object -ExpandProperty BusType | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($diskType)) {
		$arr['disk_type'] = [System.String]$diskType.Trim().ToLower()
	} else {
		Write-Host "Disk type not found. Setting disk_type to null."
		$arr['disk_type'] = $null
	}
} catch {
	Write-Host "Error retrieving disk type: $_. Setting disk_type to null."
	$arr['disk_type'] = $null
}

# Disk size in KB
$arr['disk_size_kb'] = $null
try {
	$diskSizeBytes = ($win32DiskDriveObj | Measure-Object -Property Size -Sum).Sum
	if ($null -ne $diskSizeBytes -and [System.Int64]$diskSizeBytes -gt 0) {
		$arr['disk_size_kb'] = [System.Int64]($diskSizeBytes / 1024)
	} else {
		Write-Host "Disk size not found. Setting disk_size_kb to null."
		$arr['disk_size_kb'] = $null
	}
} catch {
	Write-Host "Error retrieving disk size: $_. Setting disk_size_kb to null."
	$arr['disk_size_kb'] = $null
}

# Disk free space in KB
$arr['disk_free_space_kb'] = $null
try {
	$diskFreeSpaceBytes = ($win32LogicalDiskObj | Measure-Object -Property FreeSpace -Sum).Sum
	if ($null -ne $diskFreeSpaceBytes -and [System.Int64]$diskFreeSpaceBytes -ge 0) {
		$arr['disk_free_space_kb'] = [System.Int64]($diskFreeSpaceBytes / 1024)
	} else {
		Write-Host "Disk free space not found. Setting disk_free_space_kb to null."
		$arr['disk_free_space_kb'] = $null
	}
} catch {
	Write-Host "Error retrieving disk free space: $_. Setting disk_free_space_kb to null."
	$arr['disk_free_space_kb'] = $null
}

# Ethernet MAC address
$arr['ethernet_mac_addr'] = $null
try {
	$ethernetMac = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($ethernetMac)) {
		$arr['ethernet_mac_addr'] = [System.String]$ethernetMac.Trim().Replace("-", ":")
	} else {
		Write-Host "Ethernet MAC address not found. Setting ethernet_mac_addr to null."
		$arr['ethernet_mac_addr'] = $null
	}
} catch {
	Write-Host "Error retrieving Ethernet MAC address: $_. Setting ethernet_mac_addr to null."
	$arr['ethernet_mac_addr'] = $null
}

# Wi-Fi MAC address
$arr['wifi_mac_addr'] = $null
# Interface type 71 is for wireless interfaces
$wifiInterface = Get-NetAdapter -Physical | Where-Object { $_.Status -eq "Up" -and $_.InterfaceType -eq 71 } | Select-Object -First 1
if ($null -ne $wifiInterface) {
	$wifiMac = ($wifiInterface | Select-Object -ExpandProperty MacAddress)
	if (-not [System.String]::IsNullOrWhiteSpace($wifiMac)) {
		$arr['wifi_mac_addr'] = [System.String]$wifiMac.Trim().Replace("-", ":")
	} else {
		Write-Host "Wi-Fi MAC address not found."
	}
} else {
	Write-Host "Wi-Fi interface not found."
}

# Battery manufacturer
$arr['battery_manufacturer'] = $null
$arr['battery_serial'] = $null
$arr['battery_current_max_capacity'] = $null
$arr['battery_design_capacity'] = $null
$arr['battery_charge_cycles'] = $null
$arr['battery_health'] = $null
# Win32_Battery class
if ($null -ne $win32BatteryObj) {

	# Battery static data class
	if ($null -ne $batteryStaticDataObj) {
		# Battery manufacturer
		if (-not [System.String]::IsNullOrWhiteSpace($batteryStaticDataObj.ManufactureName)) {
			$arr['battery_manufacturer'] = [System.String]$batteryStaticDataObj.ManufactureName.Trim()
		} else {
			Write-Host "Battery manufacturer not found."
		}
		# Battery fully charged design capacity
		if ($null -ne $batteryStaticDataObj.DesignedCapacity -and [System.Int64]$batteryStaticDataObj.DesignedCapacity -gt 0) {
			$arr['battery_design_capacity'] = [System.Int64]$batteryStaticDataObj.DesignedCapacity
		} else {
			Write-Host "Battery design capacity not found."
		}
	} else {
		Write-Host "BatteryStaticData WMI class not found."
	}

	# Battery current max capacity class
	if ($null -ne $batteryFullChargedCapacityObj) {
		$batteryCurrentMaxCapacityRaw = $batteryFullChargedCapacityObj.FullChargedCapacity
		$batteryCurrentMaxCapacity = [System.Int64]0
		if ([System.Int64]::TryParse($batteryCurrentMaxCapacityRaw, [ref]$batteryCurrentMaxCapacity) -and $batteryCurrentMaxCapacity -gt 0) {
			$arr['battery_current_max_capacity'] = $batteryCurrentMaxCapacity
		} else {
			Write-Host "Cannot parse battery current max capacity."
		}
	} else {
		Write-Host "BatteryFullChargedCapacity WMI class not found."
	}

	# Battery cycle count class
	if ($null -ne $batteryCycleCountObj) {
		$batteryCycleCountRaw = $batteryCycleCountObj.CycleCount
		$batteryCycleCount = [System.Int64]0
		if ([System.Int64]::TryParse($batteryCycleCountRaw, [ref]$batteryCycleCount) -and $batteryCycleCount -ge 0) {
			$arr['battery_charge_cycles'] = $batteryCycleCount
		} else {
			Write-Host "Cannot parse battery cycle count."
		}
	} else {
		Write-Host "BatteryCycleCount WMI class not found."
	}

	# Battery health calculation
	$batteryHealth = $null
	if ($null -ne $arr['battery_design_capacity'] -and [System.Int64]$arr['battery_design_capacity'] -gt 0 -and $null -ne $arr['battery_current_max_capacity'] -and [System.Int64]$arr['battery_current_max_capacity'] -gt 0) {
		$batteryHealth = ([System.Double]$arr['battery_current_max_capacity'] / [System.Double]$arr['battery_design_capacity']) * 100
		$arr['battery_health'] = [System.Double]$batteryHealth
	} else {
		Write-Host "Battery health cannot be calculated due to missing design capacity or current max capacity."
	}
} else {
	Write-Host "Battery not found."
}

# Updated from Windows boolean
$arr['updated_from_windows'] = [System.Boolean]$true

$httpBodyArr = @{}
foreach ($key in $arr.Keys) {
	if ($null -ne $arr[$key]) {
		$httpBodyArr[$key] = $arr[$key]
	}
	if ($null -eq $arr[$key]) {
		$httpBodyArr[$key] = $null
	}
	if ($arr[$key] -is [System.String]) {
		$httpBodyArr[$key] = $arr[$key].Trim()
	}
	if ($arr[$key] -is [System.String] -and [System.String]::IsNullOrWhiteSpace($arr[$key])) {
		$httpBodyArr[$key] = $null
	}
	if ($key -eq "ethernet_mac_addr" -and $null -ne $httpBodyArr[$key]) {
		$httpBodyArr[$key] = $httpBodyArr[$key].Replace("-", ":")
	}
	if ($key -eq "wifi_mac_addr" -and $null -ne $httpBodyArr[$key]) {
		$httpBodyArr[$key] = $httpBodyArr[$key].Replace("-", ":")
	}
}

Add-Type -AssemblyName System.Windows.Forms
$dialogObj = New-Object System.Windows.Forms.FolderBrowserDialog
$dialogObj.Description = "Select a folder to save the backups"
$fileDialog = $dialogObj.ShowDialog()

if ($fileDialog -eq [System.Windows.Forms.DialogResult]::OK -and 
	-not [System.String]::IsNullOrWhiteSpace($dialogObj.SelectedPath)) {
	Write-Host "Selected folder: $($dialogObj.SelectedPath)"
} else {
	Write-Host "No folder selected. Exiting."
	exit
}

$jsonStr = $httpBodyArr | ConvertTo-Json -Depth 4
Out-File -FilePath "$($dialogObj.SelectedPath)\uit-system-info.json" -InputObject $jsonStr -Encoding UTF8
Write-Host $jsonStr