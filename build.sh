echo "Building binaries"

GOOS=linux GOARCH=amd64 go build -o bin/janitorv0.1.0_linux_amd64
GOOS=windows GOARCH=amd64 go build -o bin/janitorv0.1.0_windows_amd64.exe
GOOS=darwin GOARCH=arm64 go build -o bin/janitorv0.1.0_darwin_arm64