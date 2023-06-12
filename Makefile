TESTS := $(wildcard tests/*.sh)

build: 
	go build

test: build
	$(MAKE) $(TESTS)
	@printf '\e[32m Passed all tests \e[0m\n'

$(TESTS):
	@echo 'Testing' $@
	@./$@
	@printf '\e[32m OK \e[0m\n'

clean:
	go clean
	rm -rf out/

.PHONY: build clean test $(TESTS)