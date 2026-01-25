
add:
	go run \
		./cmd/cli \
			feed add \
				--feed "https://medium.com/@eldius/feed" \
				--feed "https://dev.to/feed/eldius" \
				--feed "https://dev.to/feed/tag/go" \
				--feed "https://dev.to/feed/pachicodes" \
				--feed "https://www.asemanago.dev/feed" \
				--max-results 20


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
					search "how to debug golang from command line"


ask:
	go run \
        		./cmd/cli \
        			ask \
        				O que é um MCP no contexto de LLMs? Me dê exemplos de uso.

release:
	goreleaser \
		release \
			--clean \
			--snapshot

testing:
	go run ./cmd/cli testing