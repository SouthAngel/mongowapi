@echo off
chcp 65001 >nul
echo ========================================
echo   MongoWAPI Release Build
echo ========================================

echo.
echo [1/2] Building Windows amd64...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o release\mongowapi-windows-amd64.exe .
if errorlevel 1 (
    echo [FAIL] Windows build failed
    exit /b 1
)
echo [OK]   release\mongowapi-windows-amd64.exe

echo.
echo [2/2] Building Linux amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o release\mongowapi-linux-amd64 .
if errorlevel 1 (
    echo [FAIL] Linux build failed
    exit /b 1
)
echo [OK]   release\mongowapi-linux-amd64

echo.
echo ========================================
echo   Build completed successfully
echo ========================================
dir /B release
