#!/bin/zsh 

GOARCH=amd64 GOOS=linux go build -o main

zip main.zip main

aws lambda update-function-code \
  --function-name agent \
  --zip-file fileb://main.zip | jq

