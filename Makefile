## build: Build the application
.PHONY: build
build: compile
	@go build -o ./tmp/main ./cmd/api/


.PHONY: compile
compile:
	@echo -n '** Generating tailwind.css file | '
	@npx @tailwindcss/cli -i ./static/css/input.css -o ./static/css/tailwind.css
	@echo -n '** Generating templ | '
	templ generate

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

templ:
	@templ generate --watch --proxy="http://localhost:8090" --open-browser=false -v




# # Run templ generation in watch mode
# templ:
# 	templ generate --watch --proxy="http://localhost:8090" --open-browser=false -v
#
# ## Run air for Go hot reload
# server:
# 		air \
# 	    --build.cmd "go build -o ./tmp/main ./cmd/api/" \
# 	    --build.bin "./tmp/main" \
# 	    --build.delay "100" \
# 	    --build.exclude_dir "node_modules" \
# 	    --build.include_ext "go" \
# 	    --build.stop_on_error "false" \
# 	    --misc.clean_on_exit true
#
# # Watch Tailwind CSS changes
# tailwind:
# 	@tailwindcss/cli -i ./ui/static/css/input.css -o ./ui/static/css/tailwind.css --watch
#
# # Start development server with all watchers
# dev:
# 	make -j3 tailwind templ server
