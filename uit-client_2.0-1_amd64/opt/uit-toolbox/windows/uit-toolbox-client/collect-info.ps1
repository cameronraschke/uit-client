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


$computerInfoObj = (Get-ComputerInfo)
$win32CompSysObj = (Get-CimInstance -Class Win32_ComputerSystem)
$win32ComputerSystemProductObj = (Get-CimInstance -Class Win32_ComputerSystemProduct)
$win32BiosObj = (Get-CimInstance -Class Win32_BIOS)
# $win32OperatingSystemObj = (Get-CimInstance -Class Win32_OperatingSystem)
$win32MemoryObj = (Get-CimInstance -Class Win32_PhysicalMemory)
$win32ProcessorObj = (Get-CimInstance -Class Win32_Processor)
$win32DiskDriveObj = (Get-CimInstance -Class Win32_DiskDrive -Filter "MediaType != 'Removable Media' AND interfaceType = 'SCSI'")
$win32LogicalDiskObj = (Get-CimInstance -Class Win32_LogicalDisk -Filter "DriveType = '3' AND Name = 'C:'")
$win32BatteryObj = (Get-CimInstance -Class Win32_Battery -ErrorAction SilentlyContinue)
$batteryStaticDataObj = (Get-WmiObject -Namespace "root\wmi" -Class "BatteryStaticData" -ErrorAction SilentlyContinue)
$batteryCycleCountObj = (Get-WmiObject -Namespace "root\wmi" -ClassName BatteryCycleCount -ErrorAction SilentlyContinue)
$batteryFullChargedCapacityObj = (Get-CimInstance -Namespace "root\wmi" -ClassName "BatteryFullChargedCapacity" -ErrorAction SilentlyContinue)
$dsregObj = (dsregcmd /status)
$installedPackages = Get-Package -ErrorAction SilentlyContinue

$jsonObject = [PSCustomObject]@{
	request_metadata = @{
		timestamp = [System.DateTime]::Now.ToString("yyyy-MM-dd'T'HH:mm:sszzz")
		transaction_uuid = $null #[System.Guid]::NewGuid().ToString()
		updated_from_windows = $true
		tagnumber = [System.Int64]$tagNum
		system_serial = $null
	}
	system_uuid = $null
	system_manufacturer = $null
	system_model = $null
	system_sku = $null
	chassis_type = $null
	bios_version = $null
	bios_release_date = $null
	tpm_version = $null
	secure_boot_enabled = $null
	os_installed_at = $null
	os_vendor = $null
	os_platform = $null
	os_architecture = $null
	os_name = $null
	os_version = $null
	windows_display_version = $null
	windows_build_number = $null
	windows_ubr = $null
	windows_bitlocker_enabled = $null
	ad_domain = $null
	computer_name = $null
	ad_computer_name = $null
	ad_distinguished_name = $null
	ad_admin_users = $null
	is_intune_joined = $null
	memory_capacity_kb = $null
	memory_serial = $null
	memory_speed_mhz = $null
	cpu_model = $null
	cpu_core_count = $null
	cpu_thread_count = $null
	disk_model = $null
	disk_type = $null
	disk_size_kb = $null
	disk_free_space_kb = $null
	ethernet_mac_addr = $null
	wifi_mac_addr = $null
	battery_manufacture_date = $null
	battery_manufacturer = $null
	battery_model = $null
	battery_serial = $null
	battery_current_max_capacity = $null
	battery_design_capacity = $null
	battery_charge_cycles = $null
	battery_health = $null
	installed_apps = $null
	has_2023_ca = $false
}

# System UUID/SMBIOS GUID
$jsonObject.system_uuid = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32ComputerSystemProductObj.UUID)) {
	$jsonObject.system_uuid = [System.String]$win32ComputerSystemProductObj.UUID.Trim()
} else {
	Write-Host "System SMBIOS GUID not found in WMI."
}

# System serial
$jsonObject.request_metadata.system_serial = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32ComputerSystemProductObj.IdentifyingNumber)) {
	$jsonObject.request_metadata.system_serial = [System.String]$win32ComputerSystemProductObj.IdentifyingNumber.Trim()
} elseif (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.BiosSeralNumber)) {
	# This is misspelled in PowerShell
	$jsonObject.request_metadata.system_serial = [System.String]$computerInfoObj.BiosSeralNumber.Trim()
} else {
	Write-Host "System serial number not found in WMI."
}

# System manufacturer
$jsonObject.system_manufacturer = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32CompSysObj.Manufacturer)) {
	$jsonObject.system_manufacturer = [System.String]$win32CompSysObj.Manufacturer.Trim()
} else {
	Write-Host "System manufacturer not found in WMI."
}

# System model
$jsonObject.system_model = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32CompSysObj.Model)) {
	$jsonObject.system_model = [System.String]$win32CompSysObj.Model.Trim()
} else {
	Write-Host "System model not found in WMI."
}

# System SKU
$jsonObject.system_sku = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32CompSysObj.SystemSKUNumber)) {
	$jsonObject.system_sku = [System.String]$win32CompSysObj.SystemSKUNumber.Trim()
} else {
	Write-Host "System SKU not found in WMI."
}

# Chassis type
$jsonObject.chassis_type = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsPCSystemType)) {
	$jsonObject.chassis_type = ([System.String]$computerInfoObj.CsPCSystemType).Trim()
} else {
	Write-Host "Chassis type not found in WMI."
}

# BIOS version
$jsonObject.bios_version = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32BiosObj.SMBIOSBIOSVersion)) {
	$jsonObject.bios_version = [System.String]$win32BiosObj.SMBIOSBIOSVersion.Trim()
} else {
	Write-Host "BIOS version not found in WMI."
}

# BIOS release date
$jsonObject.bios_release_date = $null
if (-not [System.String]::IsNullOrWhiteSpace($win32BiosObj.ReleaseDate)) {
	$parsedBiosDate = [System.DateTime]::MinValue
	if ([System.DateTime]::TryParse($win32BiosObj.ReleaseDate, [ref]$parsedBiosDate)) {
		$jsonObject.bios_release_date = [System.String]$parsedBiosDate.ToString("yyyy-MM-dd'T'HH:mm:sszzzz")
	} else {
		Write-Host "Failed to parse BIOS release date."
		$jsonObject.bios_release_date = $null
	}
} else {
	Write-Host "BIOS release date not found in WMI."
}

#TPM version
$jsonObject.tpm_version = $null
try {
	$tpmVersion = (Get-WmiObject -Namespace "Root\CIMv2\Security\MicrosoftTpm" -Class Win32_Tpm | Select-Object -ExpandProperty SpecVersion) -split ", " | Select-Object -First 1
	if (-not [System.String]::IsNullOrWhiteSpace($tpmVersion)) {
		$jsonObject.tpm_version = [System.String]$tpmVersion
	} else {
		Write-Host "TPM version not found in WMI."
	}
} catch {
	Write-Host "Error retrieving TPM version: $_"
}

# Secure boot state
$jsonObject.secure_boot_enabled = $null
try {
	$secureBootEnabled = Confirm-SecureBootUEFI -ErrorAction SilentlyContinue
	if ($null -ne $secureBootEnabled) {
		$jsonObject.secure_boot_enabled = [System.Boolean]$secureBootEnabled
	} else {
		Write-Host "Secure boot state not found."
	}
} catch {
	Write-Host "Error retrieving secure boot state: $_"
}

# OS Install date
$jsonObject.os_installed_at = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OsInstallDate)) {
	$parsedOSInstallDate = [System.DateTime]::MinValue
	if ([System.DateTime]::TryParse($computerInfoObj.OsInstallDate, [ref]$parsedOSInstallDate)) {
		$jsonObject.os_installed_at = [System.String]$parsedOSInstallDate.ToString("yyyy-MM-dd'T'HH:mm:sszzzz")
	} else {
		Write-Host "Failed to parse OS install date."
	}
} else {
	Write-Host "OS install date not found in WMI."
}

# OS vendor
$jsonObject.os_vendor = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSManufacturer)) {
	$jsonObject.os_vendor = [System.String]$computerInfoObj.OSManufacturer.Trim()
} else {
	Write-Host "OS vendor not found in WMI."
}

# OS platform
$jsonObject.os_platform = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OsType)) {
	$jsonObject.os_platform = ([System.String]$computerInfoObj.OsType).Trim()
} else {
	Write-Host "OS platform not found in WMI."
}

# OS architecture
$jsonObject.os_architecture = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OsArchitecture)) {
	$jsonObject.os_architecture = [System.String]$computerInfoObj.OsArchitecture.Trim()
} else {
	Write-Host "OS architecture not found in WMI."
}

# OS name
$jsonObject.os_name = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSName)) {
	$jsonObject.os_name = [System.String]$computerInfoObj.OSName.Trim()
} else {
	Write-Host "OS name not found in WMI."
}

# OS version
$jsonObject.os_version = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSVersion)) {
	$jsonObject.os_version = [System.String]$computerInfoObj.OSVersion.Trim()
} else {
	Write-Host "OS version not found in WMI."
}

# Windows display version
$jsonObject.windows_display_version = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.OSDisplayVersion)) {
	$jsonObject.windows_display_version = [System.String]$computerInfoObj.OSDisplayVersion.Trim()
} else {
	Write-Host "Windows display version not found in WMI."
}

# Windows build number
$jsonObject.windows_build_number = $null
$windowsBuildNumberRaw = $computerInfoObj.OsBuildNumber
if (-not [System.String]::IsNullOrWhiteSpace($windowsBuildNumberRaw)) {
	$windowsBuildNumber = [System.Int64]0
	if ([System.Int64]::TryParse([string]$windowsBuildNumberRaw, [ref]$windowsBuildNumber) -and $windowsBuildNumber -gt 0) {
		$jsonObject.windows_build_number = $windowsBuildNumber
	} else {
		Write-Host "Windows build number not found in WMI."
	}
} else {
	Write-Host "Windows build number not found in WMI."
}

# Windows UBR
$jsonObject.windows_ubr = $null
try {
	$ubrValue = (Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name UBR).UBR
	if ([System.Int64]::TryParse($ubrValue, [ref]$null)) {
		$jsonObject.windows_ubr = [System.Int64]$ubrValue
	} else {
		Write-Host "Windows UBR value not found in registry."
	}
} catch {
	Write-Host "Error retrieving Windows UBR value from registry: $_"
}

# Windows BitLocker enabled
$jsonObject.windows_bitlocker_enabled = $null
try {
	$bitlockerStatus = (Get-BitLockerVolume -MountPoint "C:").VolumeStatus
	if (-not [System.String]::IsNullOrWhiteSpace($bitlockerStatus)) {
		$jsonObject.windows_bitlocker_enabled = [System.Boolean]($bitlockerStatus -eq "FullyEncrypted")
	} else {
		Write-Host "BitLocker status not found."
	}
} catch {
	Write-Host "Error retrieving BitLocker status: $_"
}

# AD domain
$jsonObject.ad_domain = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsDomain)) {
	$jsonObject.ad_domain = [System.String]($computerInfoObj.CsDomain).Trim()
} else {
	Write-Host "AD domain not found in WMI."
}

# computer name
$jsonObject.computer_name = $null
if (-not [System.String]::IsNullOrWhiteSpace($computerInfoObj.CsDNSHostName)) {
	$jsonObject.computer_name = [System.String]($computerInfoObj.CsDNSHostName).Trim()
} else {
	Write-Host "AD domain computer name not found in WMI."
}

#AD domain computer name
$jsonObject.ad_computer_name = $null
($dsregObj | Out-String -Stream | Select-String -Pattern "Device Name" | ForEach-Object { $_ -replace '\s', '' } | ForEach-Object { $_ -replace '^.*:', '' }) | ForEach-Object {
	if (-not [System.String]::IsNullOrWhiteSpace($_)) {
		$jsonObject.ad_computer_name = [System.String]$_
	} else {
		Write-Host "AD domain computer name not found from dsregcmd output."
	}
}


# AD distinguished name
$jsonObject.ad_distinguished_name = $null
Get-ADComputer -Identity $env:COMPUTERNAME -Properties DistinguishedName -ErrorAction SilentlyContinue | ForEach-Object {
	if (-not [System.String]::IsNullOrWhiteSpace($_.DistinguishedName)) {
		$jsonObject.ad_distinguished_name = [System.String]$_.DistinguishedName.Trim()
	} else {
		Write-Host "AD distinguished name not found for computer $env:COMPUTERNAME."
	}
}

# List of AD admin users
$jsonObject.ad_admin_users = $null
$jsonObject.ad_admin_users = (Get-LocalGroupMember -Group "Administrators" -ErrorAction SilentlyContinue | Where-Object { ($_.ObjectClass -eq "User") -and ($_.PrincipalSource -eq "ActiveDirectory") } | Select-Object -ExpandProperty Name | Sort-Object) -join ";"

# Is intune joined 
$jsonObject.is_intune_joined = $null
try {
	$isAzureJoined = (($dsregObj | Out-String -Stream | Select-String -Pattern "AzureAdJoined" | ForEach-Object { $_ -replace '\s', '' } | ForEach-Object { $_ -replace '^.*:', '' }) -eq "YES")
	$isDomainJoined = (($dsregObj | Out-String -Stream | Select-String -Pattern "DomainJoined" | ForEach-Object { $_ -replace '\s', '' } | ForEach-Object { $_ -replace '^.*:', '' }) -eq "YES")
	$isIntuneJoined = $isAzureJoined -and $isDomainJoined	
		$jsonObject.is_intune_joined = [System.Boolean]($isIntuneJoined)
} catch {
	Write-Host "Error determining Intune join status: $_"
}

# Memory capacity in KB
$jsonObject.memory_capacity_kb = $null
try {
	$memoryCapacityBytes = ($win32MemoryObj | Measure-Object -Property Capacity -Sum).Sum
	if ($null -ne $memoryCapacityBytes -and [System.Int64]$memoryCapacityBytes -gt 0) {
		$jsonObject.memory_capacity_kb = [System.Int64]($memoryCapacityBytes / 1024)
	} else {
		Write-Host "Memory capacity not found."
	}
} catch {
	Write-Host "Error retrieving memory capacity: $_"
}

# Memory serial numbers (per DIMM)
$jsonObject.memory_serial = $null
try {
	$memorySerialNumbers = $win32MemoryObj | Select-Object -ExpandProperty SerialNumber | Where-Object { -not [System.String]::IsNullOrWhiteSpace($_) } | ForEach-Object { $_.Trim() }
	if ($memorySerialNumbers.Count -gt 0) {
		$jsonObject.memory_serial = $memorySerialNumbers -join ";"
	} else {
		Write-Host "Memory serial numbers not found."
	}
} catch {
	Write-Host "Error retrieving memory serial numbers: $_"
}

# memory speed in MHz
$jsonObject.memory_speed_mhz = $null
try {
	$memorySpeedRaw = ($win32MemoryObj | Select-Object -ExpandProperty Speed | Select-Object -First 1)
	$memorySpeed = [System.Int64]0
	if ([System.Int64]::TryParse([string]$memorySpeedRaw, [ref]$memorySpeed) -and $memorySpeed -gt 0) {
		$jsonObject.memory_speed_mhz = $memorySpeed
	} else {
		Write-Host "Memory speed not found."
	}
} catch {
	Write-Host "Error retrieving memory speed: $_"
}

# CPU model
$jsonObject.cpu_model = $null
try {
	$cpuModel = ($win32ProcessorObj | Select-Object -ExpandProperty Name | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($cpuModel)) {
		$jsonObject.cpu_model = [System.String]$cpuModel.Trim()
	} else {
		Write-Host "CPU model not found."
	}
} catch {
	Write-Host "Error retrieving CPU model: $_"
}

# CPU core count
$jsonObject.cpu_core_count = $null
try {
	$cpuCoreCountRaw = ($win32ProcessorObj | Select-Object -ExpandProperty NumberOfCores | Select-Object -First 1)
	$cpuCoreCount = [System.Int64]0
	if ([System.Int64]::TryParse([string]$cpuCoreCountRaw, [ref]$cpuCoreCount) -and $cpuCoreCount -gt 0) {
		$jsonObject.cpu_core_count = $cpuCoreCount
	} else {
		Write-Host "CPU core count not found."
	}
} catch {
	Write-Host "Error retrieving CPU core count: $_"
}

# CPU thread count
$jsonObject.cpu_thread_count = $null
try {
	$cpuThreadCountRaw = ($win32ProcessorObj | Select-Object -ExpandProperty NumberOfLogicalProcessors | Select-Object -First 1)
	$cpuThreadCount = [System.Int64]0
	if ([System.Int64]::TryParse([string]$cpuThreadCountRaw, [ref]$cpuThreadCount) -and $cpuThreadCount -gt 0) {
		$jsonObject.cpu_thread_count = $cpuThreadCount
	} else {
		Write-Host "CPU thread count not found."
	}
} catch {
	Write-Host "Error retrieving CPU thread count: $_"
}

# Disk model
$jsonObject.disk_model = $null
try {
	$diskModel = ($win32DiskDriveObj | Select-Object -ExpandProperty Model | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($diskModel)) {
		$jsonObject.disk_model = [System.String]$diskModel.Trim()
	} else {
		Write-Host "Disk model not found. Setting disk_model to null."
		$jsonObject.disk_model = $null
	}
} catch {
	Write-Host "Error retrieving disk model: $_. Setting disk_model to null."
	$jsonObject.disk_model = $null
}

# Disk type
$jsonObject.disk_type = $null
try {
	$diskType = (Get-Disk | Where-Object { $_.DiskNumber -eq "0" } | Select-Object -ExpandProperty BusType | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($diskType)) {
		$jsonObject.disk_type = [System.String]$diskType.Trim().ToLower()
	} else {
		Write-Host "Disk type not found. Setting disk_type to null."
		$jsonObject.disk_type = $null
	}
} catch {
	Write-Host "Error retrieving disk type: $_. Setting disk_type to null."
	$jsonObject.disk_type = $null
}

# Disk size in KB
$jsonObject.disk_size_kb = $null
try {
	$diskSizeBytes = ($win32DiskDriveObj | Measure-Object -Property Size -Sum).Sum
	if ($null -ne $diskSizeBytes -and [System.Int64]$diskSizeBytes -gt 0) {
		$jsonObject.disk_size_kb = [System.Int64]($diskSizeBytes / 1024)
	} else {
		Write-Host "Disk size not found. Setting disk_size_kb to null."
		$jsonObject.disk_size_kb = $null
	}
} catch {
	Write-Host "Error retrieving disk size: $_. Setting disk_size_kb to null."
	$jsonObject.disk_size_kb = $null
}

# Disk free space in KB
$jsonObject.disk_free_space_kb = $null
try {
	$diskFreeSpaceBytes = ($win32LogicalDiskObj | Measure-Object -Property FreeSpace -Sum).Sum
	if ($null -ne $diskFreeSpaceBytes -and [System.Int64]$diskFreeSpaceBytes -ge 0) {
		$jsonObject.disk_free_space_kb = [System.Int64]($diskFreeSpaceBytes / 1024)
	} else {
		Write-Host "Disk free space not found. Setting disk_free_space_kb to null."
		$jsonObject.disk_free_space_kb = $null
	}
} catch {
	Write-Host "Error retrieving disk free space: $_. Setting disk_free_space_kb to null."
	$jsonObject.disk_free_space_kb = $null
}

# Ethernet MAC address
$jsonObject.ethernet_mac_addr = $null
try {
	$ethernetMac = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
	if (-not [System.String]::IsNullOrWhiteSpace($ethernetMac)) {
		$jsonObject.ethernet_mac_addr = [System.String]$ethernetMac.Trim().Replace("-", ":")
	} else {
		Write-Host "Ethernet MAC address not found. Setting ethernet_mac_addr to null."
		$jsonObject.ethernet_mac_addr = $null
	}
} catch {
	Write-Host "Error retrieving Ethernet MAC address: $_. Setting ethernet_mac_addr to null."
	$jsonObject.ethernet_mac_addr = $null
}

# Wi-Fi MAC address
$jsonObject.wifi_mac_addr = $null
# Interface type 71 is for wireless interfaces
$wifiInterface = Get-NetAdapter -Physical | Where-Object { $_.InterfaceType -eq 71 } | Select-Object -First 1
if ($null -ne $wifiInterface) {
	$wifiMac = ($wifiInterface | Select-Object -ExpandProperty MacAddress)
	if (-not [System.String]::IsNullOrWhiteSpace($wifiMac)) {
		$jsonObject.wifi_mac_addr = [System.String]$wifiMac.Trim().Replace("-", ":")
	} else {
		Write-Host "Wi-Fi MAC address not found."
	}
} else {
	Write-Host "Wi-Fi interface not found."
}

# Battery manufacturer
$jsonObject.battery_manufacture_date = $null
$jsonObject.battery_manufacturer = $null
$jsonObject.battery_model = $null
$jsonObject.battery_serial = $null
$jsonObject.battery_current_max_capacity = $null
$jsonObject.battery_design_capacity = $null
$jsonObject.battery_charge_cycles = $null
$jsonObject.battery_health = $null
# Win32_Battery class
if ($null -ne $win32BatteryObj) {

	# Battery static data class
	if ($null -ne $batteryStaticDataObj) {
		# Battery manufacture date
		if (-not [System.String]::IsNullOrWhiteSpace($batteryStaticDataObj.ManufactureDate)) {
			$parsedBatteryManufactureDate = [System.DateTime]::MinValue
			if ([System.DateTime]::TryParse($batteryStaticDataObj.ManufactureDate, [ref]$parsedBatteryManufactureDate)) {
				$jsonObject.battery_manufacture_date = [System.String]$parsedBatteryManufactureDate.ToString("yyyy-MM-dd'T'HH:mm:sszzzz")
			} else {
				Write-Host "Failed to parse battery manufacture date."
				$jsonObject.battery_manufacture_date = $null
			}
		} else {
			Write-Host "Battery manufacture date not found."
		}
		# Battery manufacturer
		if (-not [System.String]::IsNullOrWhiteSpace($batteryStaticDataObj.ManufactureName)) {
			$jsonObject.battery_manufacturer = [System.String]$batteryStaticDataObj.ManufactureName.Trim()
		} else {
			Write-Host "Battery manufacturer not found."
		}
		# Battery model
		if (-not [System.String]::IsNullOrWhiteSpace($batteryStaticDataObj.DeviceName)) {
			$jsonObject.battery_model = [System.String]$batteryStaticDataObj.DeviceName.Trim()
		} else {
			Write-Host "Battery model not found."
		}
		# Battery fully charged design capacity
		if ($null -ne $batteryStaticDataObj.DesignedCapacity -and [System.Int64]$batteryStaticDataObj.DesignedCapacity -gt 0) {
			$jsonObject.battery_design_capacity = [System.Int64]$batteryStaticDataObj.DesignedCapacity
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
			$jsonObject.battery_current_max_capacity = $batteryCurrentMaxCapacity
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
			$jsonObject.battery_charge_cycles = $batteryCycleCount
		} else {
			Write-Host "Cannot parse battery cycle count."
		}
	} else {
		Write-Host "BatteryCycleCount WMI class not found."
	}

	# Battery health calculation
	$batteryHealth = $null
	if ($null -ne $jsonObject.battery_design_capacity -and [System.Int64]$jsonObject.battery_design_capacity -gt 0 -and $null -ne $jsonObject.battery_current_max_capacity -and [System.Int64]$jsonObject.battery_current_max_capacity -gt 0) {
		$batteryHealth = ([System.Double]$jsonObject.battery_current_max_capacity / [System.Double]$jsonObject.battery_design_capacity) * 100
		$jsonObject.battery_health = [System.Double]$batteryHealth
	} else {
		Write-Host "Battery health cannot be calculated due to missing design capacity or current max capacity."
	}
} else {
	Write-Host "Battery not found."
}

if ($null -eq $installedPackages) {
	Write-Host "Installed applications not found."
} else {
	$jsonObject.installed_apps = ($installedPackages | Select-Object -Property Name, Version | Sort-Object -Property Name | ForEach-Object { "$($_.Name) ($($_.Version))" }) -join ";"
}

$jsonObject.has_2023_ca = $false
if ((Get-ItemPropertyValue -Path "HKLM:\SYSTEM\CurrentControlSet\Control\SecureBoot\Servicing" -Name "UEFICA2023Status") -eq "Updated" -and [System.Text.Encoding]::ASCII.GetString((Get-SecureBootUEFI db).bytes) -match 'Windows UEFI CA 2023') {
	$jsonObject.has_2023_ca = $true
}

foreach ($key in $jsonObject.PSObject.Properties.Name) {
	# Trim string values
	if ($jsonObject.$key -is [System.String]) {
		$jsonObject.$key = $jsonObject.$key.Trim()
	}
	# Set empty string values to null if empty or whitespace
	if ($jsonObject.$key -is [System.String] -and [System.String]::IsNullOrWhiteSpace($jsonObject.$key)) {
		$jsonObject.$key = $null
	}
}

$desktop = [Environment]::GetFolderPath("Desktop")
Set-Variable -Name "outDir" -Value (Join-Path $desktop "00-uit-client-system-info.json")


$jsonStr = $jsonObject | ConvertTo-Json
Out-File -FilePath "$outDir" -InputObject $jsonStr -Encoding UTF8
Write-Host $jsonStr