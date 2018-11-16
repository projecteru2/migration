deps:
	glide i
	rm -rf ./vendor/github.com/docker/docker/vendor
	rm -rf ./vendor/github.com/docker/distribution/vendor
binary:
	go build -ldflags "$(GO_LDFLAGS)" -a -tags "netgo osusergo" -installsuffix netgo -o eru-migration
