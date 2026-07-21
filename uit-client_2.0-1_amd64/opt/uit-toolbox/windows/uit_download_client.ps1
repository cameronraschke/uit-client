$desktopPath = [Environment]::GetFolderPath("Desktop")
$clientPath = Join-Path -Path $desktopPath -ChildPath "uit-toolbox-client.ps1"
curl -o $clientPath https://raw.githubusercontent.com/cameronraschke/uit-client/refs/heads/main/uit-client_2.0-1_amd64/opt/uit-toolbox/windows/uit-toolbox-client/collect-info.ps1

Start-Process -FilePath "powershell.exe" -ArgumentList "-ExecutionPolicy Bypass -File `"$clientPath`"" -Verb RunAs