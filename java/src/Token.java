/**
 * A single lexical token produced by {@link Lexer}.
 *
 * The lexeme is kept exactly as it appeared in the source so that numeric
 * literals round-trip (e.g. "5e2" stays "5e2", not "500.0").
 */
public final class Token {

    public enum Kind { IDENT, NUMBER, LPAREN, RPAREN, COMMA, EOF }

    public final Kind   kind;
    public final String lexeme;
    public final int    line;   // 1-based
    public final int    col;    // 1-based

    public Token(Kind kind, String lexeme, int line, int col) {
        this.kind   = kind;
        this.lexeme = lexeme;
        this.line   = line;
        this.col    = col;
    }

    @Override
    public String toString() {
        return kind + "('" + lexeme + "')@" + line + ":" + col;
    }
}
