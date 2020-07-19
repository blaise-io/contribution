VERSION := $(shell cat VERSION)


.PHONY: release # Create releases for popular systems.
release:
	env GOARCH=amd64 GOOS=darwin go build -ldflags="-s -w" \
		-o release/contribution-$(VERSION)-x64-macos .
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" \
		-o release/contribution-$(VERSION)-x64-linux .
	env GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" \
		-o release/contribution-$(VERSION)-x64.exe .


.PHONY: release-tag # Create and push a release tag.
release-tag:
	git tag $(VERSION)
	git push origin --tags


.PHONY: README.md # Append help command to documentation.
README.md:
	@sed '/<!-- .* -->/q' README.md > README.md.tmp
	@mv README.md.tmp README.md  # -i does not work on macOS

	@echo '### `contribution -help`\n```' >> README.md 2>&1
	go run . -help >> README.md 2>&1
	@echo '```\n' >> README.md 2>&1

	@echo '### `contribution preview -help`\n```' >> README.md 2>&1
	go run . preview -help >> README.md 2>&1
	@echo '```\n' >> README.md 2>&1

	@echo '### `contribution push -help`\n```' >> README.md 2>&1
	go run . push -help >> README.md 2>&1
	@echo '```' >> README.md 2>&1
