.PHONY: be-dev be-test fe-install fe-dev fe-test fe-build

be-dev:
	cd be && go run cmd/api/main.go
be-test:
	cd be && go test ./...
fe-install:
	cd fe && pnpm install
fe-dev:
	cd fe && pnpm dev
fe-test:
	cd fe && pnpm typecheck
fe-build:
	cd fe && pnpm build
