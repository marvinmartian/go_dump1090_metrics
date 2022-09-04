
GOARCH=arm
GOOS=linux
GOARM=6
CGO_ENABLED=false
NAME=go_dump1090_exporter

.DEFAULT_GOAL := arm

# files: src/main

vars:
	@echo "GOARCH:" $(GOARCH)
	@echo "GOOS:" $(GOOS)
	@echo "GOARM:" $(GOARM)
	@echo "CGO_ENABLED:" $(CGO_ENABLED)
	@echo "NAME:" $(NAME)
	

test:
	go test src/*.go

arm:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -tags netgo -ldflags '-w' -o $(NAME) src/metrics.go src/stats.go src/main.go

arm6:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -a -tags netgo -ldflags '-w' -o $(NAME) src/metrics.go src/stats.go src/main.go

docker:
	docker build -t $(NAME) .

docker-push: docker
	docker push $(NAME)

clean_docker:
	docker rmi $(NAME)

clean_build:
	rm $(NAME)

clean: clean_build clean_docker
	