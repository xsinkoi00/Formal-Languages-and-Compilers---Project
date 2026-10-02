# Expression Transformation Project — Java build
# Compiles all sources under src/ into out/, then runs via `java -cp out Expr`.

JAVAC  ?= javac
JAVA   ?= java
MAIN    = Expr
SRC_DIR = src
OUT_DIR = out
SOURCES = $(wildcard $(SRC_DIR)/*.java)

.PHONY: all build run test clean zip

all: build

build: $(OUT_DIR)/.compiled

$(OUT_DIR)/.compiled: $(SOURCES)
	@mkdir -p $(OUT_DIR)
	@$(JAVAC) -d $(OUT_DIR) $(SOURCES)
	@touch $(OUT_DIR)/.compiled

# Pipe an input file in:  make run IN=tests/01_example1.in
run: build
	@test -n "$(IN)" || (echo "usage: make run IN=path/to/input.in" && false)
	@$(JAVA) -cp $(OUT_DIR) $(MAIN) < $(IN)

test: build
	@bash tests/run_tests.sh

clean:
	@rm -rf $(OUT_DIR) submission.zip

# Build the submission archive. Adjust the login before submitting.
zip: clean
	@zip -r submission.zip $(SRC_DIR) Makefile README.md tests
