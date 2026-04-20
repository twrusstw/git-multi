BIN      := gitmulti
BIN_DIR  := /usr/local/bin

COMPLETION_DIR_zsh  := /usr/local/share/zsh/site-functions
COMPLETION_DIR_bash := /usr/local/share/bash-completion/completions

.PHONY: build test lint install uninstall

build:
	go build -o $(BIN) .

test:
	go test ./...

lint:
	golangci-lint run ./...

install: build
	sudo install -m 755 $(BIN) $(BIN_DIR)/$(BIN)
	sudo mkdir -p $(COMPLETION_DIR_zsh) $(COMPLETION_DIR_bash)
	sudo install -m 644 completions/_gitmulti $(COMPLETION_DIR_zsh)/_$(BIN)
	sudo install -m 644 completions/gitmulti.bash $(COMPLETION_DIR_bash)/$(BIN)
	@echo "Installed $(BIN) to $(BIN_DIR)."

uninstall:
	sudo rm -f $(BIN_DIR)/$(BIN)
	sudo rm -f $(COMPLETION_DIR_zsh)/_$(BIN)
	sudo rm -f $(COMPLETION_DIR_bash)/$(BIN)
	rm -f /tmp/gitmulti-cache-*
	@echo "Uninstall complete."
