DOCKER_REPO ?= rinq/httpd

-include artifacts/make/go.mk

.PHONY: run
run: artifacts/build/debug/$(GOOS)/$(GOARCH)/rinq-httpd
	RINQ_HTTPD_BIND=":8081" RINQ_HTTPD_ORIGIN="*" "$<"

.PHONY: run-echo-server
run-echo-server: artifacts/build/debug/$(GOOS)/$(GOARCH)/echo-server
	"$<"

artifacts/make/%.mk:
	@curl --create-dirs '-#Lo' "$@" "https://rinq.github.io/make/$*.mk?nonce=$(shell date +%s)"
