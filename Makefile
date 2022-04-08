define HELP_TEXT

  Makefile commands

	make test                     - Run the full test suite
	make start-dev-node 		  - Start a new development node

endef

help:
	$(info $(HELP_TEXT))

test:
	go test -v --cover "./..."

start-dev-node :
	 starport chain serve
