#!/bin/bash

telegram-bot-api --api-id=$TELEGRAM_API_ID --api-hash=TELEGRAM_HASH_ID --local &
go run main.go &
