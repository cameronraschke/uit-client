@echo off

echo "Select an option: "
echo "	[1] Run post-Sysprep scripts"
echo "	[2] Collect system info"
set /p userInput="Enter [1-2]: "

if "%userInput%"=="1" (
	echo "Running post-Sysprep scripts..."
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File ".\target-sysprep.ps1"
) else if "%userInput%"=="2" (
	echo "Collecting system info..."
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File ".\collect-info.ps1"
) else (
	echo "Invalid option. Exiting."
)