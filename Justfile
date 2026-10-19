default:
    @just --list

all: build

install-deps:
    @echo "Installing Go dependencies..."
    go mod tidy
    @echo "Installing Node.js dependencies..."
    bun install

frontend:
    @echo "Building frontend..."
    bun run build
    cp -a dist internal/api/

backend: frontend
    @echo "Building Go binary (pure Go, no CGO)..."
    CGO_ENABLED=0 go build -o bin/parse-dmarc .

backend-cgo: frontend
    @echo "Building Go binary (with CGO)..."
    CGO_ENABLED=1 go build -tags cgo -o bin/parse-dmarc .

build: frontend backend
    @echo "Build complete! Binary available at ./bin/parse-dmarc"

build-cgo: frontend backend-cgo
    @echo "CGO build complete! Binary available at ./bin/parse-dmarc"

dev:
    @echo "Starting development server..."
    air

config:
    go run . -gen-config

clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -rf internal/api/dist/
    rm -rf node_modules/
    rm -f config.json

test:
    go test -v ./...

install: build
    @echo "Installing to /usr/local/bin..."
    sudo cp bin/parse-dmarc /usr/local/bin/

frontend-dev:
    bun run start

update-zeabur-template:
    bunx zeabur@latest template update -f zeabur.yml -c YB3TS7
