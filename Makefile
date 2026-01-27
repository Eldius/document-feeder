
add:
	go run \
		./cmd/cli \
			feed add \
				--feed "https://medium.com/@eldius/feed" \
				--feed "https://dev.to/feed/eldius" \
				--feed "https://dev.to/feed/tag/go" \
				--feed "https://dev.to/feed/pachicodes" \
				--feed "https://www.asemanago.dev/feed" \
				--feed "https://blog.jetbrains.com/go/feed/" \
				--feed "https://blog.learngoprogramming.com/feed" \
				--feed "https://appliedgo.net/index.xml" \
				--feed "https://changelog.com/gotime/feed" \
				--feed "https://jajaldoang.com/index.xml" \
				--feed "https://gosamples.dev/index.xml" \
				--feed "https://trinhhieu668.wordpress.com/feed/" \
				--feed "https://dev.to/feed/with_code_example" \
				--feed "https://golangbot.com/index.xml"


list:
	go run \
		./cmd/cli \
			feed \
				list


refresh:
	go run \
		./cmd/cli \
			feed \
				refresh


search:
	go run \
		./cmd/cli \
			feed \
				search "golang debug from command line"


ask:
	go run \
		./cmd/cli \
			ask \
				"Explain the difference between supervised and unsupervised learning"

release:
	goreleaser \
		release \
			--clean \
			--snapshot

testing:
	go run ./cmd/cli testing

models-ls:
	go run ./cmd/cli/ models ls

models-ps:
	go run ./cmd/cli/ models ps

validate: test linter vulncheck
	@echo "Validation completed!"

test:
	go test -cover ./...

vulncheck:
	go tool govulncheck ./...

linter:
	golangci-lint run
