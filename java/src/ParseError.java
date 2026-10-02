/** Raised by the parser on a syntactically invalid program; carries source position. */
public final class ParseError extends RuntimeException {
    public final int line;
    public final int col;

    public ParseError(String msg, int line, int col) {
        super("syntax error at " + line + ":" + col + ": " + msg);
        this.line = line;
        this.col  = col;
    }
}
