be-dev:
	cd be && go run cmd/api/main.go
fe-install:
	cd fe && pnpm install
fe-dev:
	cd fe && pnpm dev
fe-build:
	cd fe && pnpm build
