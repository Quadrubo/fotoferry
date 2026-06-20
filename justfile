run *args:
    go run . {{ args }}

build:
    go build -o fotoferry .

test:
    go test ./...

lint:
    golangci-lint run

docker:
    docker compose up --build
