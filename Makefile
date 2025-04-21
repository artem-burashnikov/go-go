.PHONY: all test build clean

all: build

build:
	@$(MAKE) -C tree/ build

test:
	@$(MAKE) -C tree/ test

clean:
	@$(MAKE) -C tree/ clean
