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

Set-Variable -name "remoteHost" -Value ""
Set-Variable -name "remoteUser" -Value ""
Set-Variable -name "localDir" -Value $dialogObj.SelectedPath
Set-Variable -name "localSqlDir" -Value "$($dialogObj.SelectedPath)\sql-backups"
Set-Variable -name "localImageDir" -Value "$($dialogObj.SelectedPath)\image-backups"
Set-Variable -name "localImageZipFile" -Value "$($dialogObj.SelectedPath)\image-backups.zip"


scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/sql-backups/ $localSqlDir
scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/uit-web/inventory-images/ $localImageDir

Compress-Archive -Path "${localImageDir}\*" -DestinationPath "${localImageZipFile}" -Force
