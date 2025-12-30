@echo off
REM Migration script for Liyali Gateway (Windows)
REM Usage: migrate.bat [up|down|reset|drop]

setlocal enabledelayedexpansion

REM Check if .env file exists (go up one directory to backend root)
if not exist "../.env" (
    echo Error: .env file not found
    echo Please create a .env file with database configuration in the backend directory
    exit /b 1
)

REM Default action
set ACTION=%1
if "%ACTION%"=="" set ACTION=up

echo Liyali Gateway Database Migration
echo ==================================

if "%ACTION%"=="up" (
    echo Running UP migration...
    go run run_migration.go migrations/001_create_complete_schema.up.sql
    if !errorlevel! equ 0 (
        echo ✅ Migration completed successfully!
    ) else (
        echo ❌ Migration failed!
        exit /b 1
    )
) else if "%ACTION%"=="down" (
    echo Running DOWN migration...
    echo ⚠️  This will DROP ALL TABLES! Are you sure? (y/N)
    set /p response=
    if /i "!response!"=="y" (
        go run run_migration.go migrations/001_create_complete_schema.down.sql
        if !errorlevel! equ 0 (
            echo ✅ Rollback completed successfully!
        ) else (
            echo ❌ Rollback failed!
            exit /b 1
        )
    ) else (
        echo Migration cancelled
    )
) else if "%ACTION%"=="reset" (
    echo Resetting database (DOWN + UP)...
    echo ⚠️  This will DROP ALL TABLES and recreate them! Are you sure? (y/N)
    set /p response=
    if /i "!response!"=="y" (
        echo Step 1: Running DOWN migration...
        go run run_migration.go migrations/001_create_complete_schema.down.sql
        if !errorlevel! equ 0 (
            echo Step 2: Running UP migration...
            go run run_migration.go migrations/001_create_complete_schema.up.sql
            if !errorlevel! equ 0 (
                echo ✅ Database reset completed successfully!
            ) else (
                echo ❌ UP migration failed!
                exit /b 1
            )
        ) else (
            echo ❌ DOWN migration failed!
            exit /b 1
        )
    ) else (
        echo Migration cancelled
    )
) else if "%ACTION%"=="drop" (
    echo Dropping all tables...
    echo ⚠️  This will DROP ALL TABLES! Are you sure? (y/N)
    set /p response=
    if /i "!response!"=="y" (
        go run run_migration.go migrations/000_drop_all_tables.up.sql
        if !errorlevel! equ 0 (
            echo ✅ All tables dropped successfully!
        ) else (
            echo ❌ Drop operation failed!
            exit /b 1
        )
    ) else (
        echo Operation cancelled
    )
) else (
    echo Invalid action: %ACTION%
    echo Usage: %0 [up^|down^|reset^|drop]
    echo.
    echo Actions:
    echo   up    - Run the UP migration (create tables)
    echo   down  - Run the DOWN migration (drop tables)
    echo   reset - Run DOWN then UP (complete reset)
    echo   drop  - Drop all tables using drop script
    exit /b 1
)

endlocal