UNAME := $(shell uname -s)

all:
ifeq ($(UNAME),Darwin)
	$(MAKE) -f Makefile.osx
else
	go build -v
endif

clean:
	rm -rf demo
	rm -rf Demo.app
