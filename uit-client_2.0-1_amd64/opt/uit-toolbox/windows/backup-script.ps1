# Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

$currentDate = [DateTime]::Now
$dateString = $currentDate.ToString("yyyy-MM-dd-HHmmss")

$remoteHost = ""
if ($remoteHost -eq "") {
		Write-Host "Error: Remote host is not specified. Please set the remote host variable."
		exit 1
}
$remoteUser = ""
if ($remoteUser -eq "") {
		Write-Host "Error: Remote user is not specified. Please set the remote user variable."
		exit 1
}
$desktopDir = [Environment]::GetFolderPath("Desktop")
$tmpDir = Join-Path $env:TEMP "${dateString}-uit-backup"
$pgDumpDir = Join-Path $tmpDir "sql-backups"
$oldImagesDir = Join-Path $tmpDir "image-backups"
$migratedImagesDir = Join-Path $tmpDir "migrated-image-backups"
$outFile = Join-Path $desktopDir "00-uit-web-backups\${dateString}-uit-backup.zip"

mkdir -Path $tmpDir -Force
mkdir -Path $pgDumpDir -Force
mkdir -Path $oldImagesDir -Force
mkdir -Path $migratedImagesDir -Force

scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/sql-backups/* $pgDumpDir
scp -r ${remoteUser}@${remoteHost}:/opt/uit-toolbox/uit-web/inventory-images/* $oldImagesDir
scp -r ${remoteUser}@${remoteHost}:/opt/inventory_images/* $migratedImagesDir

Compress-Archive -Path "${tmpDir}\*" -DestinationPath "${outFile}" -Force

if (Test-Path $outFile) {
	Write-Host "Backup completed successfully. Backup archive path: '${outFile}'"
	if (Test-Path $tmpDir) {
		Remove-Item -Path $tmpDir -Recurse -Force
	}
} else {
	Write-Host "Backup failed. Temporary files at '${tmpDir}' are being retained."
}
