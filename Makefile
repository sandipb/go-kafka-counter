.PHONY: build builddockerbuild buildindocker clean


build:
	go build ./cmd/kafka-count

builddockerbuild:
	docker-compose build appbuild

buildindocker:
	docker-compose run --rm -- appbuild

clean:
	rm -f kafka-count