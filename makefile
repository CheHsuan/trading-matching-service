.PHONY: run rund test install-swag clean

run:
	swag init
	go run ./main.go

rund: clean
	docker-compose up --build -d
	docker-compose logs -f server

test:
	go test -v ./test/unittest

install-swag:
	go get -u github.com/swaggo/swag/cmd/swag

clean:
	docker-compose down --remove-orphans
