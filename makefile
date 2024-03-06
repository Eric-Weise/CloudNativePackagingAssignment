SHELL := /bin/bash

.PHONY: help
help:
	@echo "Usage make <TARGET>"
	@echo ""
	@echo "  Targets:"
	@echo "	   build				Build the voter executable"
	@echo "	   run					Run the voter program from code"
	@echo "	   run-bin				Run the voter executable"
	@echo "	   load-db				Add sample data via curl"
	@echo "	   get-by-id			Get a voter by id pass id=<id> on command line"
	@echo "	   get-all				Get all voters"
	@echo "	   update-2				Update record 2, pass a new title in using title=<title> on command line"
	@echo "	   delete-all			Delete all voters"
	@echo "	   delete-by-id			Delete a voter by id pass id=<id> on command line"
	@echo "	   get-v2				Get all voters by done status pass done=<true|false> on command line"
	@echo "	   get-v2-all			Get all voters using version 2"
	@echo "	   build-amd64-linux	Build amd64/Linux executable"
	@echo "	   build-arm64-linux	Build arm64/Linux executable"





.PHONY: build
build:
	go build .

.PHONY: build-amd64-linux
build-amd64-linux:
	GOOS=linux GOARCH=amd64 go build -o ./voter-linux-amd64 .

.PHONY: build-arm64-linux
build-arm64-linux:
	GOOS=linux GOARCH=arm64 go build -o ./voter-linux-arm64 .

	
.PHONY: run
run:
	go run main.go

.PHONY: run-bin
run-bin:
	./voter

.PHONY: restore-db
restore-db:
	(cp ./data/voter.json.bak ./data/voter.json)

.PHONY: restore-db-windows
restore-db-windows:
	(copy.\data\voter.json.bak .\data\voter.json)

.PHONY: load-db
load-db:
	curl -d '{ "id": 1, "title": "Learn Go / GoLang", "done": false }' -H "Content-Type: application/json" -X POST http://localhost:1080/voter 
	curl -d '{ "id": 2, "title": "Learn Kubernetes", "done": true}' -H "Content-Type: application/json" -X POST http://localhost:1080/voter 
	curl -d '{ "id": 3, "title": "Learn Cloud Native Architecture","done": false}' -H "Content-Type: application/json" -X POST http://localhost:1080/voter 
	curl -d '{"id": 4,"title": "Learn Why Professor Mitchell is the BEST! :-)","done": true}' -H "Content-Type: application/json" -X POST http://localhost:1080/voter

.PHONY: update-2
update-2:
	curl -d '{ "id": 2, "title": "$(title)", "done": false }' -H "Content-Type: application/json" -X PUT http://localhost:1080/voter 

.PHONY: get-by-id
get-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voter/$(id) 

.PHONY: get-all
get-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/voter 

.PHONY: delete-all
delete-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/voter 

.PHONY: delete-by-id
delete-by-id:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X DELETE http://localhost:1080/voter/$(id) 

.PHONY: get-v2
get-v2:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/v2/voter?done=$(done) 

.PHONY: get-v2-all
get-v2-all:
	curl -w "HTTP Status: %{http_code}\n" -H "Content-Type: application/json" -X GET http://localhost:1080/v2/voter
