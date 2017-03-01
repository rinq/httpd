-include artifacts/build/Makefile.in

.PHONY: run
run: $(BUILD_PATH)/debug/$(CURRENT_OS)/$(CURRENT_ARCH)/rinq-httpd
	$(BUILD_PATH)/debug/$(CURRENT_OS)/$(CURRENT_ARCH)/rinq-httpd

artifacts/build/Makefile.in:
	mkdir -p "$(@D)"
	curl -Lo "$(@D)/runtime.go" https://raw.githubusercontent.com/icecave/make/master/go/runtime.go
	curl -Lo "$@" https://raw.githubusercontent.com/icecave/make/master/go/Makefile.in
