@echo off

set bat_dir=%~dp0
cd %bat_dir%

go run ./cmd --data ..\data\geodata --level district --addr :8080