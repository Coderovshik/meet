APP_NAME = meet
DOCKER_IMAGE = meet
PORT = 8080

.PHONY: build run docker-build docker-run clean

build:
	go build -o $(APP_NAME) ./cmd/meet/main.go

run:
	./$(APP_NAME)

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run -it --rm -p $(PORT):$(PORT) $(DOCKER_IMAGE)

clean:
	rm -f $(APP_NAME)
