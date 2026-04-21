BIN := gitmulti

XDG_DATA_HOME ?= $(HOME)/.local/share
BIN_DIR       := $(HOME)/.local/bin
ZSH_COMP_DIR  := $(XDG_DATA_HOME)/zsh/site-functions
BASH_COMP_DIR := $(XDG_DATA_HOME)/bash-completion/completions

SHELL_NAME := $(shell basename $${SHELL:-sh})

# BSD (macOS) vs GNU sed in-place flag
SED_IN_PLACE := $(shell sed --version 2>/dev/null | grep -q GNU && echo '-i' || echo "-i ''")

.PHONY: build test lint install uninstall doctor

build:
	go build -o $(BIN) .

test:
	go test ./...

lint:
	golangci-lint run ./...

install: build
	install -d $(BIN_DIR) $(ZSH_COMP_DIR) $(BASH_COMP_DIR)
	install -m 755 $(BIN) $(BIN_DIR)/$(BIN)
	install -m 644 _gitmulti $(ZSH_COMP_DIR)/_gitmulti
	install -m 644 gitmulti.bash $(BASH_COMP_DIR)/gitmulti
	rm -f $(HOME)/.zcompdump*
	@echo ""
	@echo "Installed (user-level, no sudo):"
	@echo "  bin  -> $(BIN_DIR)/$(BIN)"
	@echo "  zsh  -> $(ZSH_COMP_DIR)/_gitmulti"
	@echo "  bash -> $(BASH_COMP_DIR)/gitmulti"
	@$(MAKE) --no-print-directory doctor

doctor:
	@ok=1; \
	case ":$$PATH:" in \
		*:$(BIN_DIR):*) ;; \
		*) ok=0; \
		   echo ""; \
		   echo "[PATH] $(BIN_DIR) not in PATH. Add to your shell rc:"; \
		   echo "    export PATH=\"$(BIN_DIR):\$$PATH\"" ;; \
	esac; \
	if [ "$(SHELL_NAME)" = "zsh" ]; then \
		if ! zsh -ic 'print -l $$fpath' 2>/dev/null | grep -qx "$(ZSH_COMP_DIR)"; then \
			ok=0; \
			echo ""; \
			echo "[zsh] $(ZSH_COMP_DIR) not in \$$fpath. Add to ~/.zshrc BEFORE compinit:"; \
			echo "    fpath=($(ZSH_COMP_DIR) \$$fpath)"; \
			echo "    autoload -Uz compinit && compinit"; \
		fi; \
	fi; \
	if [ "$(SHELL_NAME)" = "bash" ]; then \
		if ! bash -ic 'type _init_completion' >/dev/null 2>&1; then \
			ok=0; \
			echo ""; \
			echo "[bash] bash-completion@2 not active. Install and enable:"; \
			echo "    # macOS: brew install bash-completion@2"; \
			echo "    # Linux: apt/dnf install bash-completion"; \
			echo "    # then ensure ~/.bashrc sources its bash_completion.sh"; \
			echo "    # user-level completions under $(BASH_COMP_DIR) are auto-loaded once active"; \
		fi; \
	fi; \
	if [ $$ok -eq 1 ]; then \
		echo ""; \
		echo "Shell prerequisites OK ($(SHELL_NAME)). Restart your shell to activate."; \
	fi

uninstall:
	rm -f $(BIN_DIR)/$(BIN)
	rm -f $(ZSH_COMP_DIR)/_gitmulti
	rm -f $(BASH_COMP_DIR)/gitmulti
	rm -rf $(HOME)/.git-multi
	@for rc in $(HOME)/.zshrc $(HOME)/.bashrc $(HOME)/.profile; do \
		if [ -f "$$rc" ] && grep -q "gitmulti completion" "$$rc" 2>/dev/null; then \
			sed $(SED_IN_PLACE) '/# Add gitmulti completion/d' "$$rc"; \
			sed $(SED_IN_PLACE) '/source.*auto-completion\.sh/d' "$$rc"; \
			echo "Removed legacy rc entries from $$rc."; \
		fi; \
	done
	rm -f /tmp/gitmulti-cache-* $(HOME)/.zcompdump*
	@echo "Uninstall complete."
