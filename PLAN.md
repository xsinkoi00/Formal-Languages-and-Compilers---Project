# Expression Transformation Project — Build Plan

**Language:** Python 3
**Deadline:** 2026-12-06 23:59 (StudIS)
**Target server:** merlin.fit.vutbr.cz

Transforms a prefix function composition (e.g. `add(5, mul(3, sub(10, pow(6, 4))))`)
into an equivalent infix math expression (`5 + 3 * (10 - 6^4)`) with the minimum
number of parentheses.

---

## Phase 0 — Project scaffolding

- Decide file layout. Proposed:
  - `expr.py` — single-file entry point (reads stdin, prints result).
  - `tests/` — folder of `.in` / `.out` pairs.
  - `run_tests.sh` or `make test` — compares each `.in` output against its `.out`.
  - `Makefile` — `make`, `make test`, `make clean`.
  - `README.md` — authorship, build/run on Merlin, FA diagram, notes.
- Confirm Python version available on Merlin (3.x). No third-party libraries.

**Exit criteria:** empty `expr.py` runs on Merlin via `python3 expr.py`.

---

## Phase 1 — Lexical analyzer (scanner)

Implemented as an **explicit finite automaton** (needed for the README diagram).

### Tokens
| Token      | Lexeme                                   |
|------------|------------------------------------------|
| `IDENT`    | `add`, `sub`, `mul`, `div`, `mod`, `pow`, `tern` |
| `NUMBER`   | per grammar: `[+-]?digits(.digits)?(e[+-]?digits)?` |
| `LPAREN`   | `(`                                      |
| `RPAREN`   | `)`                                      |
| `COMMA`    | `,`                                      |
| `EOF`      | end of input                             |

Skipped: whitespace, `/* ... */` block comments (non-nested).

### FA states (sketch)
- `S0` start — dispatch on first char.
- `S_ID` — reading identifier letters.
- `S_NUM_SIGN` → `S_NUM_INT` → `S_NUM_DOT` → `S_NUM_FRAC` → `S_NUM_E` → `S_NUM_ESIGN` → `S_NUM_EXP`.
- `S_COMMENT_SLASH` → `S_COMMENT_BODY` → `S_COMMENT_STAR` → back to `S0`.
- Final states emit a token and return to `S0`.

### Deliverables
- `Lexer` class with `next_token()` / `peek()`.
- Clear error messages with position (line, column).
- Hand-drawn FA diagram saved as `fa.png` (or Mermaid in README).

**Exit criteria:** running the lexer on each sample input prints the correct token stream.

---

## Phase 2 — Parser + AST

Grammar (input language — pure prefix):

```
expr    ::= NUMBER
          | IDENT '(' expr ',' expr [',' expr] ')'
```

- Recursive descent, one function `parse_expr()`.
- Validate arity: `tern` → 3 args; `add|sub|mul|div|mod|pow` → 2 args.
- Reject unknown identifiers.

### AST nodes
```python
Num(text)                 # keep original lexeme so "5e2" round-trips
Op(symbol, children, prec, right_assoc)
```
where `symbol` is one of `+ - * / % ^ ?:` and `prec` / `right_assoc` are set
from a fixed table.

### Operator table
| Func   | Symbol | Precedence | Associativity |
|--------|--------|------------|---------------|
| `tern` | `?:`   | 1          | right         |
| `add`  | `+`    | 2          | left          |
| `sub`  | `-`    | 2          | left          |
| `mul`  | `*`    | 3          | left          |
| `div`  | `/`    | 3          | left          |
| `mod`  | `%`    | 3          | left          |
| `pow`  | `^`    | 4          | right         |

**Exit criteria:** parser builds a correct AST for each sample input; syntax errors reported with position.

---

## Phase 3 — Pretty-printer (minimum-parentheses infix)

The core intellectual task.

For every child of a binary operator parent:
- Let `pp` = parent precedence, `pc` = child precedence.
- Wrap the child in `(...)` **iff**:
  - `pc < pp`, **or**
  - `pc == pp` **and** the child sits on the *wrong* side for the parent's associativity
    (left child of a right-assoc parent, or right child of a left-assoc parent).

Ternary (`?:`) printing: `cond ? then : else`, right-associative, lowest precedence.
A `Num` child never needs parentheses. A negative-number lexeme like `-5` printed next to `-` or in an exponent position may need special-casing — list this as a known edge case.

**Exit criteria:** all four example outputs match exactly:
1. `add(5, mul(3, sub(10, pow(6, 4))))` → `5 + 3 * (10 - 6^4)`
2. `pow(pow(2,3),1)` → `(2^3)^1`
3. `tern(5e2, 2.7,1)` → `5e2?2.7:1`
4. `tern(1,tern(2, 3, 4) ,5)` → `1?2?3:4:5`

Note the spacing: examples 1 shows spaces around `+`, `*`, `-`; examples 2–4 show no spaces. Decide a consistent style and justify it in the README (safe default: no spaces, matches 3 of 4 examples).

---

## Phase 4 — Driver / I/O

- Read **all of stdin** (expression can span lines/have arbitrary whitespace).
- Run lexer → parser → printer.
- Print result + newline to stdout.
- Exit codes: `0` OK, non-zero on lexical/syntactic error (document the chosen codes in README).

---

## Phase 5 — Testing

- At minimum the four assignment examples as `.in` / `.out` pairs.
- Extra cases to add:
  - Deeply nested `pow` both ways (`pow(2,pow(3,4))` vs `pow(pow(2,3),4)`).
  - `sub(sub(1,2),3)` vs `sub(1,sub(2,3))` — left-assoc parens behavior.
  - Numbers with every optional part: `-1.5e+3`, `+2.`, bare `0`.
  - Comment handling: `add(/*x*/1,2)`.
  - Whitespace: newlines inside the expression.
  - Error cases: wrong arity, unknown function, unterminated comment, stray chars.
- `make test` iterates and diffs.

---

## Phase 6 — README & packaging

README (≤ 2 A4) sections:
1. Author (name, login).
2. Build/run on Merlin (`make`, `python3 expr.py < input.txt`).
3. Finite-automaton diagram (image or Mermaid).
4. Implementation notes: Python 3, stdlib only; design of parser & printer.
5. Known bugs / unimplemented parts.
6. AI / internet references (list anything used, including this planning session).

Archive contents: `expr.py`, `tests/`, `Makefile`, `README.md`, `fa.png`.
Verify everything runs on Merlin before submitting.

---

## Risk list
- **Parenthesization edge cases** around unary signs inside numbers (`sub(1,-2)` printed as `1 - -2`?). Decide a rule and test it.
- **Right-associativity for ternary** when nested on the condition side: `tern(tern(a,b,c),d,e)` → needs the left child wrapped. Covered by the standard rule above.
- **Merlin Python version** — stick to features available in Python 3.6+ to be safe.
