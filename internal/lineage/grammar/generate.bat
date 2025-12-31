@echo off
REM Generate Go code from ANTLR grammar files

set SCRIPT_DIR=%~dp0
set OUTPUT_DIR=%SCRIPT_DIR%..\parser

REM Create output directory if not exists
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM Generate Go code
antlr4 -Dlanguage=Go -visitor -listener -package parser -o "%OUTPUT_DIR%" "%SCRIPT_DIR%SQLLexer.g4" "%SCRIPT_DIR%SQLParser.g4"

echo Generated Go parser code in %OUTPUT_DIR%
