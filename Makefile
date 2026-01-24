
add:
	go run \
		./cmd/cli \
			feed add \
				--feed "https://medium.com/@eldius/feed" \
				--feed "https://dev.to/feed/eldius" \
				--feed "https://dev.to/feed/tag/go" \
				--feed "https://dev.to/feed/pachicodes"


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


release:
	goreleaser \
		release \
			--clean \
			--snapshot
