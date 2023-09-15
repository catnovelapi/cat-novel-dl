@echo off
chcp 65001 > nul

setlocal enabledelayedexpansion

set /p account=please input account:
set /p token=please input token:

echo "1: input 1 to download book by id"
echo "2: input 2 to search book and download"
set /p inputValue=please input a number:

if %inputValue% == 1 (
    set /p bookId="please input a bookId:"
    start go run main.go -d !bookId! -a %account% -t %token%
) else if %inputValue% == 2 (
    set /p bookName="please input a bookName:"
    start go run main.go -s !bookName! -a %account% -t %token%
) else   (
    echo "input error"
)