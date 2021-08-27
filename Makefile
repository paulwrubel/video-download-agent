.PHONY: all run build

all: build run

build: clean
	docker-compose build

run: clean build
	docker-compose up -d

logs: run
	docker-compose logs

follow: run
	docker-compose logs -f

clean:
	docker-compose down