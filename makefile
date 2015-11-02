GO_FILES = $(wildcard *.go) $(wildcard **/*.go)

it: $(GO_FILES)
	go build -o it

clean:
	rm it

clean-debug:
	rm tmp/built*
