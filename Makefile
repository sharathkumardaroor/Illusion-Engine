.PHONY: all engine ui clean test

# Detect OS
ifeq ($(OS),Windows_NT)
    EXE_EXT := .exe
    RM := del /Q
    MKDIR := mkdir
else
    EXE_EXT :=
    RM := rm -f
    MKDIR := mkdir -p
endif

ENGINE_BIN := chronos-engine$(EXE_EXT)

all: engine ui

engine:
	cd engine && go build -o $(ENGINE_BIN) main.go

# Using bundle instead of linux/windows because of environment limitations
ui:
	cd ui && flutter build bundle

test:
	cd ui && flutter test
	cd engine && go test ./...

clean:
	$(RM) engine/$(ENGINE_BIN)
	cd ui && flutter clean
