
REMOTE_USER := eldius
REMOTE_HOST := 192.168.0.152
REMOTE_DIR := .bin/docs/bin
REMOTE_OS := linux
REMOTE_ARCH := arm64_v8.0

push: release
	REMOTE_USER=$(REMOTE_USER) \
		REMOTE_HOST=$(REMOTE_HOST) \
		REMOTE_DIR=$(REMOTE_DIR) \
		REMOTE_OS=$(REMOTE_OS) \
		REMOTE_ARCH=$(REMOTE_ARCH) \
			./scripts/push.sh

remote-benchmark: push
	ssh $(REMOTE_USER)@$(REMOTE_HOST) 'cd ./.bin/docs/bin; ./benchmarker-cli  --model=tinyllama:latest --model=deepseek-r1:7b'

add:
	go run \
		./cmd/cli \
			feed add \
				--feed "https://cloudwithazeem.medium.com/feed" \
				--feed "https://alexrios.me/rss.xml" \
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
				--feed "https://golangbot.com/index.xml" \
				--feed "https://cprss.s3.amazonaws.com/golangweekly.com.xml" \
				--feed "https://codingplainenglish.medium.com/feed" \
				--feed "https://levelup.gitconnected.com/feed" \
				--feed "https://medium.com/@samurai.stateless.coder/feed" \
				--feed "https://blog.stackademic.com/feed" \
				--feed "https://medium.com/@grinaldiwisnu/feed" \
				--feed "https://medium.com/@utsavmadaan823/feed" \
				--feed "https://medium.com/@aleksei.aleinikov.gr/feed" \
				--feed "https://medium.com/@gopinathr143/feed" \
				--feed "https://gauravsarma1992.medium.com/feed" \
				--feed "https://medium.com/@yardenlaif/feed" \
				--feed "https://medium.com/@briannqc/feed" \
				--feed "https://medium.com/@jamal.kaksouri/feed" \
				--feed "https://medium.com/@ravikumar19997/feed" \
				--feed "https://medium.com/@sanilkhurana7/feed" \
				--feed "https://arshad404.medium.com//feed" \
				--feed "https://computaria.gitlab.io/blog/feed.xml" \
				--feed "https://www.dolthub.com/blog/rss.xml" \
				--feed "https://blog.gaborkoos.com/feed.xml" \
				--feed "https://antonz.org/feed.xml" \
				--feed "https://pt.globalvoices.org/feed/" \
				--feed "http://agenciabrasil.ebc.com.br/rss/ultimasnoticias/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/ultimasnoticias/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/ultimasnoticias/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/direitos-humanos/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/economia/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/educacao/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/esportes/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/geral/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/internacional/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/justica/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/politica/feed.xml" \
				--feed "http://agenciabrasil.ebc.com.br/rss/saude/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/rss/ultimasnoticias/parceiros/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/ultimasnoticias/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/cultura/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/direitos-humanos/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/economia/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/educacao/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/esportes/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/geral/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/internacional/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/justica/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/meio-ambiente/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/pesquisa-e-inovacao/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/politica/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/saude/feed.xml" \
				--feed "https://agenciabrasil.ebc.com.br/radioagencia-nacional/rss/seguranca/feed.xml"

list:
	go run \
		./cmd/cli \
			feed \
				list


refresh:
	go run \
		./cmd/cli \
			feed \
				refresh -i


search:
	go run \
		./cmd/cli \
			feed \
				search \
					"explain how to debug a golang code from command line"


ask:
	go run \
		./cmd/cli \
			ask \
				"Explain the difference between supervised and unsupervised learning"


sanitize:
	go run \
		./cmd/cli/ \
			sanitize

release:
	go tool \
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

models-autoconf:
	go run ./cmd/cli/ \
		models \
		autoconfigure \
			--model=all-minilm

benchmark:
	go run ./cmd/benchmarker \
		--model=tinyllama:latest \
		--model=deepseek-r1:7b

benchmark-show:
	go run ./cmd/benchmarker show \
		--model=tinyllama:latest \
		--model=deepseek-r1:7b

validate: test linter vulncheck
	@echo "Validation completed!"

test:
	go test ./... -coverage

vulncheck:
	go tool govulncheck ./...

linter:
	golangci-lint run

serve:
	go run ./cmd/web

api-add:
	curl -N -i -XPOST "http://localhost:8080/api/feeds" \
		-H "Content-Type: application/json" \
		-d @docs/samples/add_feeds_payload.json
