-include artifacts/make/go.mk

# .PHONY: run
# run: $(BUILD_PATH)/debug/$(CURRENT_OS)/$(CURRENT_ARCH)/rinq-httpd
# 	RINQ_BIND=":8081" RINQ_ORIGIN="*" "$<"
#
# .PHONY: run-echo-server
# run-echo-server: $(BUILD_PATH)/debug/$(CURRENT_OS)/$(CURRENT_ARCH)/echo-server
# 	"$<"

artifacts/make/%.mk:
	@curl --create-dirs '-#Lo' "$@" "https://rinq.github.io/make/$*.mk?nonce=$(shell date +%s)"
