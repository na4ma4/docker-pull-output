MATRIX_OS ?= darwin linux windows
MATRIX_ARCH ?= amd64 386

-include artifacts/make/go/Makefile

artifacts/make/%/Makefile:
	curl -sf https://jmalloc.github.io/makefiles/fetch | bash /dev/stdin $*
