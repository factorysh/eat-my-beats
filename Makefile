build: bin
	go build -o bin/eat-my-beats .

bin:
	mkdir -p bin

linux: bin
	make build GOOS=linux GOARCH=amd64 CGO_ENABLED=0
	upx bin/eat-my-beats
