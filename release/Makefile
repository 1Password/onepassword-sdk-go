export SDK_NAME=go
export REPO_NAME=onepassword-sdk-${SDK_NAME}

export SDK_VERSION = v$(shell cat .version)
export SDK_BUILD_NUMBER = "$(shell cat .version-build)"
export SDK_BUILD_FILE := $(CURDIR)/.version-build

prep-go:
	release-scripts/go-release.sh

.PHONY:version_swap

version_swap:
	cd version_swap/src && cargo run -r main.rs

sdk-release-go:
	release-scripts/sdk-release.sh