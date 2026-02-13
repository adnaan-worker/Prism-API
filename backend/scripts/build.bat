@echo off
REM 构建脚本 - Windows 版本

setlocal enabledelayedexpansion

REM 获取版本信息
if "%VERSION%"=="" set VERSION=1.0.0

REM 获取当前时间（ISO 8601 格式）
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set BUILD_TIME=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2%T%datetime:~8,2%:%datetime:~10,2%:%datetime:~12,2%Z

REM 获取 Git 提交哈希
for /f %%i in ('git rev-parse HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

REM 输出目录
set OUTPUT_DIR=bin
if not exist %OUTPUT_DIR% mkdir %OUTPUT_DIR%

REM 构建标志
set LDFLAGS=-X "api-aggregator/backend/pkg/utils.Version=%VERSION%" -X "api-aggregator/backend/pkg/utils.BuildTime=%BUILD_TIME%" -X "api-aggregator/backend/pkg/utils.GitCommit=%GIT_COMMIT%"

echo Building Prism API...
echo   Version: %VERSION%
echo   Build Time: %BUILD_TIME%
echo   Git Commit: %GIT_COMMIT%
echo.

REM 构建服务器
echo Building server...
go build -ldflags "%LDFLAGS%" -o %OUTPUT_DIR%\server.exe .\cmd\server

REM 构建迁移工具
echo Building migrate tool...
go build -ldflags "%LDFLAGS%" -o %OUTPUT_DIR%\migrate.exe .\scripts\migrate.go

echo.
echo Build completed successfully!
echo Binaries are in %OUTPUT_DIR%\

endlocal
