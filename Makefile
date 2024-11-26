## build: Build the application
.PHONY: build
build: compile
	@go build -o ./tmp/main ./cmd/web/


.PHONY: compile
compile:
	@echo -n '** Generating tailwind.css file | '
	@npx tailwindcss -i ./ui/static/css/input.css -o ./ui/static/css/tailwind.css --minify

## run: Run the binary
.PHONY: run
run: build
	@./tmp/web



## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code
