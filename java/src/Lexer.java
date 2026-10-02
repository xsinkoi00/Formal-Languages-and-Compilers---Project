import java.util.ArrayList;
import java.util.List;

/**
 * Explicit deterministic finite automaton for the project's lexical grammar.
 *
 * See the state diagram in README.md. Accepting (token-emitting) states are:
 *   S_ID            -> IDENT
 *   S_NUM_INT       -> NUMBER
 *   S_NUM_FRAC      -> NUMBER
 *   S_NUM_EXP       -> NUMBER
 *   single-char states for '(' ')' ','
 *
 * Whitespace and block comments ("/* ... *&#47;") are consumed silently in
 * state S0 and never produce a token.
 *
 * The automaton is encoded as a method ({@link #nextToken()}) that drives a
 * state variable until a complete token is produced, keeping the mapping to
 * the diagram one-to-one.
 */
public final class Lexer {

    private final String src;
    private int i    = 0;
    private int line = 1;
    private int col  = 1;

    public Lexer(String source) {
        this.src = source;
    }

    // -----------------------------------------------------------------
    // Public API
    // -----------------------------------------------------------------

    /** Lex the entire input into a token list (ending with an EOF token). */
    public List<Token> tokenize() {
        List<Token> out = new ArrayList<>();
        while (true) {
            Token t = nextToken();
            out.add(t);
            if (t.kind == Token.Kind.EOF) return out;
        }
    }

    /** Drive the FA until a single complete token is produced. */
    public Token nextToken() {
        while (true) {
            // S0: skip whitespace.
            while (hasMore() && isSpace(peek(0))) advance();

            int startLine = line;
            int startCol  = col;

            if (!hasMore()) {
                return new Token(Token.Kind.EOF, "", startLine, startCol);
            }

            char c = peek(0);

            // Single-character punctuation — direct final transitions.
            if (c == '(') { advance(); return new Token(Token.Kind.LPAREN, "(", startLine, startCol); }
            if (c == ')') { advance(); return new Token(Token.Kind.RPAREN, ")", startLine, startCol); }
            if (c == ',') { advance(); return new Token(Token.Kind.COMMA,  ",", startLine, startCol); }

            // Block comment: S0 --'/'--> S_SLASH --'*'--> S_COMMENT ...
            if (c == '/') {
                consumeComment(startLine, startCol);
                continue;   // back to S0
            }

            // Identifier: S0 --letter--> S_ID
            if (isAlpha(c)) {
                return scanIdent(startLine, startCol);
            }

            // Number: S0 --sign|digit--> S_NUM_SIGN | S_NUM_INT
            if (c == '+' || c == '-' || isDigit(c)) {
                return scanNumber(startLine, startCol);
            }

            throw new LexError("unexpected character '" + c + "'", startLine, startCol);
        }
    }

    // -----------------------------------------------------------------
    // Cursor helpers
    // -----------------------------------------------------------------

    private boolean hasMore() { return i < src.length(); }

    private char peek(int offset) {
        int j = i + offset;
        return (j < src.length()) ? src.charAt(j) : '\0';
    }

    private char advance() {
        if (i >= src.length()) return '\0';
        char c = src.charAt(i++);
        if (c == '\n') { line++; col = 1; }
        else           { col++; }
        return c;
    }

    private static boolean isSpace(char c) {
        return c == ' ' || c == '\t' || c == '\r' || c == '\n';
    }

    private static boolean isDigit(char c) { return c >= '0' && c <= '9'; }

    // Identifiers in this language are only the seven function names, all
    // lowercase ASCII letters. Keep the character class tight — anything
    // else is a lexical error.
    private static boolean isAlpha(char c) { return c >= 'a' && c <= 'z'; }

    // -----------------------------------------------------------------
    // Sub-automata
    // -----------------------------------------------------------------

    private void consumeComment(int startLine, int startCol) {
        advance();                              // S0 -> S_SLASH
        if (peek(0) != '*') {
            throw new LexError("stray '/' (did you mean '/* ... */'?)",
                    startLine, startCol);
        }
        advance();                              // S_SLASH -> S_COMMENT

        // S_COMMENT loops; '*' moves to S_COMMENT_STAR; from there '/' closes
        // the comment, '*' stays, anything else falls back to S_COMMENT.
        while (true) {
            if (!hasMore()) {
                throw new LexError("unterminated block comment", startLine, startCol);
            }
            char ch = advance();
            if (ch == '*') {
                while (peek(0) == '*') advance();
                if (peek(0) == '/') { advance(); return; }
            }
        }
    }

    private Token scanIdent(int startLine, int startCol) {
        StringBuilder sb = new StringBuilder();
        while (hasMore() && isAlpha(peek(0))) sb.append(advance());
        // The lexer only validates the character class; verifying that the
        // name is one of {add, sub, mul, div, mod, pow, tern} is the parser's
        // job (it gives better errors with context).
        return new Token(Token.Kind.IDENT, sb.toString(), startLine, startCol);
    }

    private Token scanNumber(int startLine, int startCol) {
        StringBuilder sb = new StringBuilder();
        char c = peek(0);

        // S_NUM_SIGN: optional leading '+'/'-' MUST be glued to a digit.
        if (c == '+' || c == '-') {
            sb.append(advance());
            if (!isDigit(peek(0))) {
                throw new LexError("expected digit after sign", line, col);
            }
        }

        // S_NUM_INT: integer digits.
        while (isDigit(peek(0))) sb.append(advance());

        // Optional fraction: '.' <digit-sequence>  (dot requires >=1 digit).
        if (peek(0) == '.') {
            sb.append(advance());               // S_NUM_INT -> S_NUM_DOT
            if (!isDigit(peek(0))) {
                throw new LexError("expected digit after '.'", line, col);
            }
            while (isDigit(peek(0))) sb.append(advance());   // S_NUM_FRAC
        }

        // Optional exponent: 'e' <optional-sign> <digit-sequence>.
        // Grammar is lowercase-only, so 'E' is rejected.
        if (peek(0) == 'e') {
            sb.append(advance());               // -> S_NUM_E
            if (peek(0) == '+' || peek(0) == '-') {
                sb.append(advance());           // S_NUM_E -> S_NUM_ESIGN
            }
            if (!isDigit(peek(0))) {
                throw new LexError("expected digit in exponent", line, col);
            }
            while (isDigit(peek(0))) sb.append(advance());   // S_NUM_EXP
        }

        return new Token(Token.Kind.NUMBER, sb.toString(), startLine, startCol);
    }
}
