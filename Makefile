DOCKER_REPO ?= rinq/httpd

SHELL := /bin/bash
-include artifacts/make/go.mk

.PHONY: run
run: artifacts/build/debug/$(GOOS)/$(GOARCH)/rinq-httpd
	RINQ_HTTPD_BIND=":8081" RINQ_HTTPD_ORIGIN="*" "$<"

artifacts/make/%.mk:
	bash <(curl -s https://rinq.github.io/make/install) $*
