.PHONY: build test vet clean docker

build:
	go build -o slack-export .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f slack-export

docker:
	docker build -t dcaslin/slack-export .
