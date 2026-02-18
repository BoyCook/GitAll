GOBIN := $(shell go env GOPATH)/bin

.PHONY: install test vet build

install: build
	go install .
	@if ! echo "$$PATH" | grep -q "$(GOBIN)"; then \
		case "$$SHELL" in \
			*/zsh)  SHELL_RC="$$HOME/.zshrc" ;; \
			*/bash) SHELL_RC="$$HOME/.bashrc" ;; \
			*)      SHELL_RC="$$HOME/.profile" ;; \
		esac; \
		echo 'export PATH="$$HOME/go/bin:$$PATH"' >> "$$SHELL_RC"; \
		echo "Added $(GOBIN) to PATH in $$SHELL_RC — restart your terminal or run: source $$SHELL_RC"; \
	else \
		echo "Installed gitall to $(GOBIN)"; \
	fi

build:
	go build -o /dev/null .

test:
	go test -race -count=1 ./...

vet:
	go vet ./...
