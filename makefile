all: test
	go build -o build/it cmd/it/inca-tool.go

clean:
	rm -f it

clean-debug: clean
	rm -f tmp/built*

test:
	go test ./...

.PHONY: clean clean-debug test
