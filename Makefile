send: 
	go run cmd/send/main.go

.PHONY: send

receive: 
	go run cmd/receive/main.go

.PHONY: receive

clean:
	rm -f *.out
.PHONY: clean

build:
	go build -o send cmd/send/main.go
	go build -o receive cmd/receive/main.go
.PHONY: build

