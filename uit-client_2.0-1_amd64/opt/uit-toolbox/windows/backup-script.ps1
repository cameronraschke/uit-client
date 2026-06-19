# Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
# Add-Type -AssemblyName System.Windows.Forms
# $dialogObj = New-Object System.Windows.Forms.FolderBrowserDialog
# $dialogObj.Description = "Select a folder to save the backups"
# $fileDialog = $dialogObj.ShowDialog()

# if ($fileDialog -eq [System.Windows.Forms.DialogResult]::OK -and -not [string]::IsNullOrWhiteSpace($dialogObj.SelectedPath)) {
# 	Write-Host "Selected folder: $($dialogObj.SelectedPath)"
# } else {
# 	Write-Host "No folder selected. Exiting."
# 	exit
# }

$currentDate = [DateTime]::Now
$dateString = $currentDate.ToString("yyyy-MM-dd-HHmmss")

Set-Variable -name "remoteHost" -Value ""
Set-Variable -name "remoteUser" -Value ""
$desktop = [Environment]::GetFolderPath("Desktop")
Set-Variable -Name "outDir" -Value (Join-Path $desktop "00-uit-web-backups\$dateString")
Set-Variable -name "localDir" -Value $dialogObj.SelectedPath
Set-Variable -name "backupPgDumpDir" -Value "$outDir\sql-backups"
Set-Variable -name "backupClientMediaDir" -Value "$outDir\image-backups"
Set-Variable -name "backupMigratedClientMediaDir" -Value "$outDir\migrated-image-backups"
# Set-Variable -name "localImageZipFile" -Value "$($dialogObj.SelectedPath)\image-backups.zip"

mkdir -Path $outDir -Force
mkdir -Path $backupPgDumpDir -Force
mkdir -Path $backupClientMediaDir -Force
mkdir -Path $backupMigratedClientMediaDir -Force

scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/sql-backups/ $backupPgDumpDir
scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/uit-web/inventory-images/ $backupClientMediaDir
scp -r ${remoteUser}@${remoteHost}:/opt/inventory_images/ $backupMigratedClientMediaDir

#Compress-Archive -Path "${backupClientMediaDir}\*" -DestinationPath "${localImageZipFile}" -Force
