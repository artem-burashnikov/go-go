.PHONY: all test build clean

all: build

build:
	@$(MAKE) -C tree/ build
	@$(MAKE) -C signer/ build

test:
	@$(MAKE) -C tree/ test
	@$(MAKE) -C signer/ test

clean:
	@$(MAKE) -C tree/ clean
	@$(MAKE) -C signer/ clean
