# Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

Set-Variable -name "remoteHost" -Value ""
Set-Variable -name "remoteUser" -Value ""
Set-Variable -name "localSqlDir" -Value "C:\Users\caraschk\Desktop\00-sql-backups"
Set-Variable -name "localImageDir" -Value "C:\Users\caraschk\Desktop\00-image-backups"
Set-Variable -name "localImageZipFile" -Value "C:\Users\caraschk\Desktop\00-image-backups.zip"


scp -r $remoteUser@$remoteHost\:/opt/uit-toolbox/sql-backups/ $localSqlDir
scp -r $remoteUser@$remoteHost\:/opt/uit-toolbox/uit-web/inventory-images/ $localImageDir

Compress-Archive -Path "$localImageDir\*" -DestinationPath "$localImageZipFile" -Force
