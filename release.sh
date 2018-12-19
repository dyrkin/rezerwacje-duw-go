#!/bin/bash

export GOARCH=amd64
export GOOS=windows
go build -o bin/rezerwacje-duw-go-win64.exe
export GOOS=linux
go build -o bin/rezerwacje-duw-go-linux64
export GOOS=darwin
go build -o bin/rezerwacje-duw-go-osx
export GOARCH=386
export GOOS=windows
go build -o bin/rezerwacje-duw-go-win32.exe
export GOOS=linux
go build -o bin/rezerwacje-duw-go-linux32
chmod +x bin/rezerwacje-duw-go-linux*
chmod +x bin/rezerwacje-duw-go-osx