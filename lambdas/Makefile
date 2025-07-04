# This Makefile orchestrates builds for all lambdas.

# Automatically discover lambda directories (directories with a Makefile)
LAMBDA_DIRS := $(shell find . -maxdepth 1 -type d -name "*" ! -name "." -exec test -f {}/Makefile \; -print | sed 's|^\./||')

# Default target
.PHONY: all build zip clean fmt vet tidy check $(LAMBDA_DIRS)

all: build

# Target to run a specific make target in all lambda directories
define run_in_lambdas
	@for dir in $(LAMBDA_DIRS); do \
		echo "Running 'make $(1)' in $$dir..."; \
		$(MAKE) -C $$dir $(1) || exit 1; \
	done
endef

build:
	$(call run_in_lambdas,build)

zip:
	$(call run_in_lambdas,zip)

clean:
	$(call run_in_lambdas,clean)

fmt:
	$(call run_in_lambdas,fmt)

vet:
	$(call run_in_lambdas,vet)

tidy:
	$(call run_in_lambdas,tidy)

check:
	$(call run_in_lambdas,check)

# Individual directory targets for convenience
$(LAMBDA_DIRS):
	@echo "Running 'make' in $@..."
	$(MAKE) -C $@

# Helper target to show discovered directories
show-dirs:
	@echo "Lambda directories: $(LAMBDA_DIRS)"


