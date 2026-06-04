# Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Add-Type -AssemblyName System.Windows.Forms
$dialogObj = New-Object System.Windows.Forms.FolderBrowserDialog
$dialogObj.Description = "Select a folder to save the backups"
$fileDialog = $dialogObj.ShowDialog()

if ($fileDialog -eq [System.Windows.Forms.DialogResult]::OK -and -not [string]::IsNullOrWhiteSpace($dialogObj.SelectedPath)) {
	Write-Host "Selected folder: $($dialogObj.SelectedPath)"
} else {
	Write-Host "No folder selected. Exiting."
	exit
}

$currentDate = [DateTime]::Now
$dateString = $currentDate.ToString("yyyy-MM-dd-HHmmss")

Set-Variable -name "remoteHost" -Value ""
Set-Variable -name "remoteUser" -Value ""
Set-Variable -name "localDir" -Value $dialogObj.SelectedPath
Set-Variable -name "localSqlDir" -Value "$($dialogObj.SelectedPath)\sql-backups\$dateString"
Set-Variable -name "localImageDir" -Value "$($dialogObj.SelectedPath)\image-backups\$dateString"
Set-Variable -name "localMigratedImageDir" -Value "$($dialogObj.SelectedPath)\migrated-image-backups\$dateString"
# Set-Variable -name "localImageZipFile" -Value "$($dialogObj.SelectedPath)\image-backups.zip"

mkdir -Path $localSqlDir -Force
mkdir -Path $localImageDir -Force
mkdir -Path $localMigratedImageDir -Force

scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/sql-backups/ $localSqlDir
scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/uit-web/inventory-images/ $localImageDir
scp -r ${remoteUser}@${remoteHost}:/opt/inventory_images/ $localMigratedImageDir

#Compress-Archive -Path "${localImageDir}\*" -DestinationPath "${localImageZipFile}" -Force
