@echo off
IF EXIST "miniJs.exe" (
	del "miniJs.exe"
)

go build