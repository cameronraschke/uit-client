# Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

Set-Variable -name "localSqlDir" -Value "C:\Users\caraschk\Desktop\00-sql-backups"
Set-Variable -name "localImageDir" -Value "C:\Users\caraschk\Desktop\00-image-backups"
Set-Variable -name "localImageZipFile" -Value "C:\Users\caraschk\Desktop\00-image-backups.zip"


scp -r cameron@172.27.53.144:/opt/uit-toolbox/sql-backups/ $localSqlDir
scp -r cameron@172.27.53.144:/opt/uit-toolbox/uit-web/inventory-images/ $localImageDir

Compress-Archive -Path "$localImageDir\*" -DestinationPath "$localImageZipFile" -Force
