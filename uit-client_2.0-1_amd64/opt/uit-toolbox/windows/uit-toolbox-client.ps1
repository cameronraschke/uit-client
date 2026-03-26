$arr = @{
'systemSerial' = (Get-CimInstance -Class Win32_ComputerSystemProduct).IdentifyingNumber
'chassisType' = (Get-CimInstance -Class Win32_ComputerSystem).ChassisSKUNumber
'adDomain' = (Get-CimInstance -Class Win32_ComputerSystem).Domain
'adDomainJoined' = (($null -ne $adDomain) -and ($adDomain -ne ""))
'systemManufacturer' = (Get-CimInstance -Class Win32_ComputerSystem).Manufacturer
'systemModel' = (Get-CimInstance -Class Win32_ComputerSystem).Model
'biosVersion' = (Get-CimInstance -Class Win32_BIOS).SMBIOSBIOSVersion
'osVersion' = (Get-CimInstance -Class Win32_OperatingSystem).BuildNumber
'osUBR' = (Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion" -Name UBR).UBR
'memorySizeBytes' = (Get-CimInstance -Class Win32_PhysicalMemory | Measure-Object -Property Capacity -Sum).Sum
'osName' = (Get-CimInstance -Class Win32_OperatingSystem).Caption
'osInstalledAt' = (Get-Date -Date ((Get-CimInstance -Class Win32_OperatingSystem).InstallDate) -UFormat '%Y-%m-%dT%H:%M:%S%Z')
'memorySpeedMHz' = (Get-CimInstance -Class Win32_PhysicalMemory).Speed | Select-Object -First 1
'cpuModel' = (Get-CimInstance -Class Win32_Processor).Name
'cpuCores' = (Get-CimInstance -Class Win32_Processor).NumberOfCores
'cpuThreads' = (Get-CimInstance -Class Win32_Processor).NumberOfLogicalProcessors
'diskSizeBytes' = (Get-CimInstance -Class Win32_DiskDrive | Measure-Object -Property Size -Sum).Sum
'ethernetMacAddr' = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
'wifiMacAddr' = (Get-CimInstance -Class Win32_NetworkAdapterConfiguration | Where-Object { $_.IPEnabled -and $_.Description -match "Wireless" } | Select-Object -ExpandProperty MACAddress | Select-Object -First 1)
'diskModel' = (Get-CimInstance -Class Win32_DiskDrive | Select-Object -ExpandProperty Model | Select-Object -First 1)
# 'cpuTemp' = (Get-CimInstance -Namespace root\wmi -Class MSAcpi_ThermalZoneTemperature | Select-Object -ExpandProperty CurrentTemperature | Select-Object -First 1) / 10 - 273.15
'batteryChargePercent' = (Get-CimInstance -Class Win32_Battery).EstimatedChargeRemaining
}

$newArr = @{}
foreach ($key in $arr.Keys) {
	if ($null -ne $arr[$key]) {
		$newArr[$key] = $arr[$key]
	}
	if ($null -eq $arr[$key]) {
		$newArr[$key] = $null
	}
	if ($arr[$key] -is [string]) {
		$newArr[$key] = $arr[$key].Trim()
	}
	if ($arr[$key] -is [string] -and [string]::IsNullOrWhiteSpace($arr[$key])) {
		$newArr[$key] = $null
	}
	if ($key -eq "ethernetMacAddr" -and $null -ne $newArr[$key]) {
		$newArr[$key] = $newArr[$key].Replace("-", ":")
	}
	if ($key -eq "wifiMacAddr" -and $null -ne $newArr[$key]) {
		$newArr[$key] = $newArr[$key].Replace("-", ":")
	}
}

$jsonStr = $newArr | ConvertTo-Json -Depth 4
Write-Host $jsonStr