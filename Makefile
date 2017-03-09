-include artifacts/build/Makefile.in

.PHONY: run
run: $(BUILD_PATH)/debug/$(CURRENT_OS)/$(CURRENT_ARCH)/rinq-httpd
	RINQ_BIND=":8081" RINQ_ORIGIN="*" "$<"

.PHONY: run-echo-server
run-echo-server: $(BUILD_PATH)/debug/$(CURRENT_OS)/$(CURRENT_ARCH)/echo-server
	"$<"

artifacts/build/Makefile.in:
	mkdir -p "$(@D)"
	curl -Lo "$(@D)/runtime.go" https://raw.githubusercontent.com/icecave/make/master/go/runtime.go
	curl -Lo "$@" https://raw.githubusercontent.com/icecave/make/master/go/Makefile.in
