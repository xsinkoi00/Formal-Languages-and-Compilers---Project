# Expression Transformation Project
# Python 3 project; nothing to compile. Targets are convenience wrappers
# so the grader can run `make test` on merlin.fit.vutbr.cz.

PYTHON ?= python3
MAIN    = expr.py

.PHONY: all run test clean zip

all:
	@echo "Nothing to build. Use: make run | make test"

# Pipe an input file in:  make run IN=tests/01_example1.in
run:
	@test -n "$(IN)" || (echo "usage: make run IN=path/to/input.in" && false)
	@$(PYTHON) $(MAIN) < $(IN)

test:
	@bash tests/run_tests.sh

clean:
	@find . -name '__pycache__' -type d -exec rm -rf {} + 2>/dev/null || true
	@rm -f submission.zip

# Build the submission archive. Adjust the login before submitting.
zip: clean
	@zip -r submission.zip $(MAIN) Makefile README.md fa.png tests
