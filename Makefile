.PHONY: all engine ui clean test

all: engine ui

engine:
	cd engine && go build -o chronos-engine main.go

# Using bundle instead of linux because gtk+-3.0 is missing in the environment
ui:
	cd ui && flutter build bundle

test:
	cd ui && flutter test
	cd engine && go test ./...

clean:
	rm -f engine/chronos-engine
	cd ui && flutter clean
