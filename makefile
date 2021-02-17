binaryName = impart-backend
imageTag = impart-backend
#.PHONY: all test clean

clean:
	@rm -f $(binaryName)

test:
	go test ./...

build: clean
	docker build -t impart-backend .

generate:
	go run cmd/modelgeneration/main.go