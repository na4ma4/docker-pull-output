MATRIX_OS ?= darwin linux windows
MATRIX_ARCH ?= amd64 386

DEBUG_ARGS   ?= -v -ldflags "-X main.version=0.0.1-debug"
RELEASE_ARGS ?= -v -ldflags "-X main.version=0.0.1 -s -w"

-include artifacts/make/go/Makefile

artifacts/make/%/Makefile:
	curl -sf https://jmalloc.github.io/makefiles/fetch | bash /dev/stdin $*
