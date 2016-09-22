GO_FILES = $(wildcard *.go) $(wildcard **/*.go)

it: $(GO_FILES)
	go build -o it

clean:
	rm -f it

clean-debug: clean
	rm -f tmp/built*
