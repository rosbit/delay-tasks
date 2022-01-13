SHELL=/bin/bash

EXE=delay-tasks

all:
	@echo "building $(EXE) ..."
	@$(MAKE) -s -f make.inc s=static

clean:
	rm -f $(EXE)
