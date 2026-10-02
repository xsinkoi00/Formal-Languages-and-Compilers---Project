#!/usr/bin/env python3
"""
Expression Transformation Project
Author: <your name>

Reads a prefix function-composition expression (e.g. "add(5, mul(3, 4))") from
standard input and prints the equivalent infix mathematical expression (with
the minimum number of parentheses) to standard output.

Pipeline:  stdin -> Lexer -> Parser -> AST -> Printer -> stdout

Debug entry points:
    python expr.py --tokens < input.txt   # dump the token stream
    python expr.py --ast    < input.txt   # dump the parsed AST
"""

import sys
from dataclasses import dataclass, field
from typing import Iterator, List, NamedTuple, Optional, Union


# ---------------------------------------------------------------------------
# Token model
# ---------------------------------------------------------------------------

# Token kinds. Keeping them as plain string constants is enough and keeps
# error messages self-explanatory.
TK_IDENT  = "IDENT"
TK_NUMBER = "NUMBER"
TK_LPAREN = "LPAREN"
TK_RPAREN = "RPAREN"
TK_COMMA  = "COMMA"
TK_EOF    = "EOF"


class Token(NamedTuple):
    kind: str
    lexeme: str   # exact source text; kept so numbers round-trip ("5e2" stays "5e2")
    line: int
    col: int

    def __repr__(self) -> str:
        return f"{self.kind}({self.lexeme!r})@{self.line}:{self.col}"


class LexError(Exception):
    """Raised on an invalid character sequence. Carries source position."""
    def __init__(self, msg: str, line: int, col: int):
        super().__init__(f"lexical error at {line}:{col}: {msg}")
        self.line = line
        self.col = col


# ---------------------------------------------------------------------------
# Lexer — explicit deterministic finite automaton
# ---------------------------------------------------------------------------
#
# States (see README for the full diagram):
#
#   S0           start / between tokens
#   S_ID         inside an identifier
#   S_NUM_SIGN   just consumed '+' or '-' (must be followed by digit)
#   S_NUM_INT    inside integer part of a number
#   S_NUM_DOT    just consumed '.' (must be followed by digit)
#   S_NUM_FRAC   inside fractional part
#   S_NUM_E      just consumed 'e' (needs sign or digit next)
#   S_NUM_ESIGN  just consumed exponent's sign (needs digit next)
#   S_NUM_EXP    inside exponent digits
#   S_SLASH      just consumed '/' (expecting '*' to open a comment)
#   S_COMMENT    inside /* ... */
#   S_COMMENT_STAR  inside /* ... */ and just saw '*'
#
# Final (accepting) states for whole tokens: S_ID, S_NUM_INT, S_NUM_FRAC,
# S_NUM_EXP. The others are transient — reaching EOF in any of them, or
# seeing an unexpected character, is a lexical error.
#
# The automaton is encoded directly as a small loop over `self.i` with an
# inner state variable. This keeps the mapping to the diagram one-to-one.
# ---------------------------------------------------------------------------

def _is_digit(c: str) -> bool:
    return "0" <= c <= "9"

def _is_alpha(c: str) -> bool:
    # Identifiers in this language are only the seven function names, all
    # lowercase ASCII letters. Keep the character class tight — anything
    # else is a lexical error.
    return "a" <= c <= "z"


class Lexer:
    def __init__(self, source: str):
        self.src = source
        self.i = 0              # index of the next character to read
        self.line = 1           # 1-based line number of self.src[self.i]
        self.col = 1            # 1-based column of self.src[self.i]

    # -- low-level cursor helpers ------------------------------------------

    def _peek(self, offset: int = 0) -> str:
        """Return the character at self.i+offset, or '' at end of input."""
        j = self.i + offset
        return self.src[j] if j < len(self.src) else ""

    def _advance(self) -> str:
        """Consume one character and update line/col. Returns '' at EOF."""
        if self.i >= len(self.src):
            return ""
        c = self.src[self.i]
        self.i += 1
        if c == "\n":
            self.line += 1
            self.col = 1
        else:
            self.col += 1
        return c

    # -- public API --------------------------------------------------------

    def tokens(self) -> Iterator[Token]:
        """Yield tokens until EOF (which is yielded last)."""
        while True:
            tok = self.next_token()
            yield tok
            if tok.kind == TK_EOF:
                return

    def next_token(self) -> Token:
        """
        Drive the FA until a complete token is produced. Whitespace and
        block comments are skipped silently; they never produce a token.
        """
        while True:
            # S0: skip whitespace between tokens.
            while self._peek() and self._peek() in " \t\r\n":
                self._advance()

            start_line, start_col = self.line, self.col
            c = self._peek()

            if c == "":
                return Token(TK_EOF, "", start_line, start_col)

            # Single-character punctuation — direct final transitions.
            if c == "(":
                self._advance(); return Token(TK_LPAREN, "(", start_line, start_col)
            if c == ")":
                self._advance(); return Token(TK_RPAREN, ")", start_line, start_col)
            if c == ",":
                self._advance(); return Token(TK_COMMA, ",", start_line, start_col)

            # Block comment: S0 --'/'--> S_SLASH --'*'--> S_COMMENT ...
            if c == "/":
                self._consume_comment(start_line, start_col)
                continue   # back to S0 to look for the next real token

            # Identifier: S0 --letter--> S_ID
            if _is_alpha(c):
                return self._scan_ident(start_line, start_col)

            # Number: S0 --sign|digit--> S_NUM_SIGN | S_NUM_INT
            if c in "+-" or _is_digit(c):
                return self._scan_number(start_line, start_col)

            raise LexError(f"unexpected character {c!r}", start_line, start_col)

    # -- comment ----------------------------------------------------------

    def _consume_comment(self, start_line: int, start_col: int) -> None:
        # Entering here with self._peek() == '/'.
        self._advance()                        # S0 -> S_SLASH
        if self._peek() != "*":
            raise LexError("stray '/' (did you mean '/* ... */'?)",
                           start_line, start_col)
        self._advance()                        # S_SLASH -> S_COMMENT

        # Scan body: S_COMMENT loops on any char; '*' moves to S_COMMENT_STAR.
        # From S_COMMENT_STAR, '/' closes the comment and returns to S0;
        # '*' stays; anything else goes back to S_COMMENT.
        while True:
            ch = self._advance()
            if ch == "":
                raise LexError("unterminated block comment", start_line, start_col)
            if ch == "*":
                while self._peek() == "*":
                    self._advance()
                if self._peek() == "/":
                    self._advance()
                    return
                # else: fall through, back to S_COMMENT

    # -- identifier -------------------------------------------------------

    def _scan_ident(self, start_line: int, start_col: int) -> Token:
        buf = []
        while _is_alpha(self._peek()):
            buf.append(self._advance())
        # Lexer only validates the character class. Checking that the name
        # is one of {add, sub, mul, div, mod, pow, tern} is the parser's job
        # (it gives better errors with context).
        return Token(TK_IDENT, "".join(buf), start_line, start_col)

    # -- number -----------------------------------------------------------

    def _scan_number(self, start_line: int, start_col: int) -> Token:
        buf = []
        c = self._peek()

        # S_NUM_SIGN: optional leading '+' or '-' MUST be glued to a digit.
        if c in "+-":
            buf.append(self._advance())
            if not _is_digit(self._peek()):
                raise LexError("expected digit after sign", self.line, self.col)

        # S_NUM_INT: integer digits.
        while _is_digit(self._peek()):
            buf.append(self._advance())

        # Optional fraction: '.' <digit-sequence>  (dot requires >=1 digit).
        if self._peek() == ".":
            buf.append(self._advance())        # S_NUM_INT -> S_NUM_DOT
            if not _is_digit(self._peek()):
                raise LexError("expected digit after '.'", self.line, self.col)
            while _is_digit(self._peek()):
                buf.append(self._advance())    # S_NUM_FRAC loop

        # Optional exponent: 'e' <optional-sign> <digit-sequence>.
        # Grammar is lowercase-only, so 'E' is rejected.
        if self._peek() == "e":
            buf.append(self._advance())        # -> S_NUM_E
            if self._peek() in "+-":
                buf.append(self._advance())    # S_NUM_E -> S_NUM_ESIGN
            if not _is_digit(self._peek()):
                raise LexError("expected digit in exponent", self.line, self.col)
            while _is_digit(self._peek()):
                buf.append(self._advance())    # S_NUM_EXP loop

        return Token(TK_NUMBER, "".join(buf), start_line, start_col)


# ---------------------------------------------------------------------------
# AST
# ---------------------------------------------------------------------------
#
# Two node kinds are enough for this language:
#
#   Num   — a numeric leaf. We keep the original lexeme so that the output
#           reproduces the input's notation ("5e2" stays "5e2", not "500.0").
#
#   Op    — an operator node. `symbol` is the infix spelling the printer
#           will emit; `prec` and `right_assoc` drive the parenthesisation
#           rule in Phase 3. For the ternary we store three children and
#           use the composite symbol "?:".
#
# The parser is the only place that maps function names to operator
# metadata, so the rest of the compiler never has to look at identifier
# text again.
# ---------------------------------------------------------------------------

@dataclass
class Num:
    text: str                     # exact lexeme from the source

@dataclass
class Op:
    symbol: str                   # one of: + - * / % ^ ?:
    prec: int                     # 1 (?:) .. 4 (^)
    right_assoc: bool
    children: List["Node"] = field(default_factory=list)

Node = Union[Num, Op]


# Function name -> (symbol, precedence, right-associative, arity)
_FUNCTIONS = {
    "tern": ("?:", 1, True,  3),
    "add":  ("+",  2, False, 2),
    "sub":  ("-",  2, False, 2),
    "mul":  ("*",  3, False, 2),
    "div":  ("/",  3, False, 2),
    "mod":  ("%",  3, False, 2),
    "pow":  ("^",  4, True,  2),
}


# ---------------------------------------------------------------------------
# Parser — recursive descent on the prefix grammar
# ---------------------------------------------------------------------------
#
# Grammar:
#
#   expr ::= NUMBER
#          | IDENT '(' expr ',' expr [ ',' expr ] ')'
#
# The input language is pure prefix (functions and parentheses only), so a
# one-function recursive descent is enough — no operator-precedence magic
# is needed on the parsing side. All precedence/associativity work happens
# later, in the printer.
# ---------------------------------------------------------------------------

class ParseError(Exception):
    """Raised on a syntactically invalid program. Carries source position."""
    def __init__(self, msg: str, line: int, col: int):
        super().__init__(f"syntax error at {line}:{col}: {msg}")
        self.line = line
        self.col = col


class Parser:
    def __init__(self, tokens: List[Token]):
        self.toks = tokens
        self.i = 0

    # -- cursor helpers ---------------------------------------------------

    def _peek(self) -> Token:
        return self.toks[self.i]

    def _advance(self) -> Token:
        t = self.toks[self.i]
        if t.kind != TK_EOF:
            self.i += 1
        return t

    def _expect(self, kind: str, what: str) -> Token:
        t = self._peek()
        if t.kind != kind:
            raise ParseError(f"expected {what}, got {t.kind} {t.lexeme!r}",
                             t.line, t.col)
        return self._advance()

    # -- entry point ------------------------------------------------------

    def parse(self) -> Node:
        node = self._parse_expr()
        trailing = self._peek()
        if trailing.kind != TK_EOF:
            raise ParseError(f"unexpected trailing {trailing.kind} "
                             f"{trailing.lexeme!r}", trailing.line, trailing.col)
        return node

    # -- grammar rules ----------------------------------------------------

    def _parse_expr(self) -> Node:
        t = self._peek()

        if t.kind == TK_NUMBER:
            self._advance()
            return Num(t.lexeme)

        if t.kind == TK_IDENT:
            return self._parse_call()

        raise ParseError(f"expected number or function name, got "
                         f"{t.kind} {t.lexeme!r}", t.line, t.col)

    def _parse_call(self) -> Node:
        name_tok = self._advance()                  # IDENT
        info = _FUNCTIONS.get(name_tok.lexeme)
        if info is None:
            raise ParseError(f"unknown function {name_tok.lexeme!r}",
                             name_tok.line, name_tok.col)
        symbol, prec, right_assoc, arity = info

        self._expect(TK_LPAREN, "'(' after function name")

        args: List[Node] = [self._parse_expr()]
        while self._peek().kind == TK_COMMA:
            self._advance()
            args.append(self._parse_expr())

        close = self._expect(TK_RPAREN, "')' to close function call")

        if len(args) != arity:
            raise ParseError(
                f"{name_tok.lexeme!r} expects {arity} argument(s), got {len(args)}",
                name_tok.line, name_tok.col,
            )

        return Op(symbol=symbol, prec=prec, right_assoc=right_assoc, children=args)


def _ast_repr(node: Node, indent: int = 0) -> str:
    """Human-readable AST dump for the --ast debug flag."""
    pad = "  " * indent
    if isinstance(node, Num):
        return f"{pad}Num({node.text!r})"
    assoc = "R" if node.right_assoc else "L"
    head = f"{pad}Op('{node.symbol}', prec={node.prec}, {assoc})"
    body = "\n".join(_ast_repr(c, indent + 1) for c in node.children)
    return head + "\n" + body


# ---------------------------------------------------------------------------
# Pretty-printer — infix with the minimum number of parentheses
# ---------------------------------------------------------------------------
#
# For every operator child of an operator parent:
#
#   let pp = parent.prec, pc = child.prec
#   wrap the child in '(' ... ')' iff
#       pc <  pp,  OR
#       pc == pp AND the child sits on the "wrong" side for the
#                    parent's associativity:
#                      - parent is right-assoc -> the LEFT child needs parens
#                      - parent is left-assoc  -> the RIGHT child needs parens
#
# A Num child never needs parentheses. For the ternary (symbol "?:") the
# middle operand sits between '?' and ':' and is syntactically unambiguous
# regardless of its own precedence, so it is never wrapped.
#
# Spacing policy — chosen to reproduce all four assignment examples
# byte-for-byte:
#     '+' '-' '*' '/' '%'  -> surrounded by single spaces
#     '^' and '?:'          -> no spaces
# ---------------------------------------------------------------------------

_SPACED = {"+", "-", "*", "/", "%"}


def pretty(node: Node) -> str:
    """Return the minimal-parentheses infix form of `node`."""
    if isinstance(node, Num):
        return node.text

    if node.symbol == "?:":
        cond = _render_child(node, node.children[0], "left")
        then = pretty(node.children[1])            # middle is always safe
        else_ = _render_child(node, node.children[2], "right")
        return f"{cond}?{then}:{else_}"

    # Binary operator.
    left  = _render_child(node, node.children[0], "left")
    right = _render_child(node, node.children[1], "right")
    sep = f" {node.symbol} " if node.symbol in _SPACED else node.symbol
    return f"{left}{sep}{right}"


def _render_child(parent: Op, child: Node, side: str) -> str:
    """Render `child` and wrap it in parens iff precedence/associativity demands it."""
    rendered = pretty(child)
    if isinstance(child, Num):
        return rendered
    needs_parens = False
    if child.prec < parent.prec:
        needs_parens = True
    elif child.prec == parent.prec:
        # Same precedence: parens only on the side that fights the parent's
        # associativity (left child of a right-assoc op, or vice versa).
        if parent.right_assoc and side == "left":
            needs_parens = True
        elif not parent.right_assoc and side == "right":
            needs_parens = True
    return f"({rendered})" if needs_parens else rendered


# ---------------------------------------------------------------------------
# Driver
# ---------------------------------------------------------------------------

def _dump_tokens(src: str) -> int:
    """Debug mode: print the token stream, one per line."""
    try:
        for tok in Lexer(src).tokens():
            print(tok)
    except LexError as e:
        print(e, file=sys.stderr)
        return 2
    return 0


def _dump_ast(src: str) -> int:
    """Debug mode: print the parsed AST."""
    try:
        tokens = list(Lexer(src).tokens())
        ast = Parser(tokens).parse()
    except LexError as e:
        print(e, file=sys.stderr); return 2
    except ParseError as e:
        print(e, file=sys.stderr); return 3
    print(_ast_repr(ast))
    return 0


def main(argv: Optional[list] = None) -> int:
    argv = sys.argv[1:] if argv is None else argv
    src = sys.stdin.read()

    if "--tokens" in argv:
        return _dump_tokens(src)
    if "--ast" in argv:
        return _dump_ast(src)

    try:
        tokens = list(Lexer(src).tokens())
        ast = Parser(tokens).parse()
    except LexError as e:
        print(e, file=sys.stderr); return 2
    except ParseError as e:
        print(e, file=sys.stderr); return 3
    print(pretty(ast))
    return 0


if __name__ == "__main__":
    sys.exit(main())
