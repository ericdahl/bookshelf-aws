# This Makefile orchestrates builds for all lambdas.

.PHONY: all build zip clean check fmt vet tidy

# Find all directories containing a Makefile (i.e., our lambda functions)
LAMBDAS := $(shell find . -mindepth 2 -maxdepth 2 -type f -name 'Makefile' | xargs -n 1 dirname)

# --- Aggregate Targets ---

all: build

build: $(foreach lambda,$(LAMBDAS),$(lambda)/build)

zip: $(foreach lambda,$(LAMBDAS),$(lambda)/zip)

clean:
	@for lambda in $(LAMBDAS); do 
		$(MAKE) -C $lambda clean; 
	done


check: $(foreach lambda,$(LAMBDAS),$(lambda)/check)

fmt: $(foreach lambda,$(LAMBDAS),$(lambda)/fmt)

vet: $(foreach lambda,$(LAMBDAS),$(lambda)/vet)

tidy: $(foreach lambda,$(LAMBDAS),$(lambda)/tidy)

# --- Dynamic Rule Generation ---

# Creates a rule for each lambda-specific target (e.g., get-book/build)
# This allows us to run 'make get-book/build' from this root Makefile
$(foreach lambda,$(LAMBDAS),$(eval $(lambda)/% : ; @$(MAKE) -C $(lambda) $*))

# Add a .PHONY rule for each generated target to ensure they always run
.PHONY: $(foreach lambda,$(LAMBDAS),$(addprefix $(lambda)/, build zip clean check fmt vet tidy))
