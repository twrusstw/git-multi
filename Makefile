BIN      := gitmulti
BIN_DIR  := /usr/local/bin
INST_DIR := $(HOME)/.git-multi

# Detect shell RC file based on $SHELL
SHELL_NAME := $(shell basename $${SHELL:-sh})
SHELL_RC    = $(HOME)/$(if $(filter zsh,$(SHELL_NAME)),.zshrc,$(if $(filter bash,$(SHELL_NAME)),.bashrc,.profile))

# BSD (macOS) vs GNU sed in-place flag
SED_INPLACE := $(shell sed --version 2>/dev/null | grep -q GNU && echo '-i' || echo "-i ''")

.PHONY: build install uninstall completion

build:
	go build -o $(BIN) .

install: build
	mkdir -p $(INST_DIR)
	sudo install -m 755 $(BIN) $(BIN_DIR)/$(BIN)
	cp auto-completion.sh $(INST_DIR)/auto-completion.sh
	chmod 644 $(INST_DIR)/auto-completion.sh
	@echo "Installed $(BIN) to $(BIN_DIR)."
	@echo "Run 'make completion' to enable tab-completion (detected RC: $(SHELL_RC))."

completion:
	@if grep -q "gitmulti completion" $(SHELL_RC) 2>/dev/null; then \
		echo "Auto-completion already present in $(SHELL_RC)."; \
	else \
		printf '\n# Add gitmulti completion\nsource $(INST_DIR)/auto-completion.sh\n' >> $(SHELL_RC); \
		echo "Auto-completion added to $(SHELL_RC). Restart your shell or run: source $(SHELL_RC)"; \
	fi

uninstall:
	sudo rm -f $(BIN_DIR)/$(BIN)
	rm -rf $(INST_DIR)
	@for rc in $(HOME)/.zshrc $(HOME)/.bashrc $(HOME)/.profile; do \
		if [ -f "$$rc" ] && grep -q "gitmulti completion" "$$rc" 2>/dev/null; then \
			sed $(SED_INPLACE) '/# Add gitmulti completion/d' "$$rc"; \
			sed $(SED_INPLACE) '/source.*auto-completion\.sh/d' "$$rc"; \
			echo "Removed auto-completion from $$rc."; \
		fi; \
	done
	@echo "Uninstall complete. Restart your shell or run: exec $$SHELL"
