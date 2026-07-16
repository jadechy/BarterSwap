.PHONY: test
test:
	docker compose exec go sh -c "go test \$$(go list ./... | grep -vE '/docs$$|/mocks$$|/database$$|^github.com/jadechy/barterswap/cmd') -v -coverprofile=/tmp/cover.out"

.PHONY: cover
cover:
	docker compose exec go sh -c "go test \$$(go list ./... | grep -vE '/docs$$|/mocks$$|/database$$|^github.com/jadechy/barterswap/cmd') -coverprofile=/tmp/cover.out && go tool cover -func=/tmp/cover.out | grep total"