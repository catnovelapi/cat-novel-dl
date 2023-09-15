#!/bin/bash

read -p "please input account: " account
read -p "please input token: " token

echo "1: 输入1以通过ID下载书籍"
echo "2: 输入2以搜索并下载书籍"
read -p "please input a number: " inputValue

if [ "$inputValue" = "1" ]; then
    read -p "please input a bookId: " bookId
    go run main.go -d "$bookId" -a "$account" -t "$token" &
elif [ "$inputValue" = "2" ]; then
    read -p "please input a bookName: " bookName
    go run main.go -s "$bookName" -a "$account" -t "$token" &
else
    echo "输入错误"
fi
