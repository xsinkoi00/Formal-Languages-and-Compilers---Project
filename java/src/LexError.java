/** Raised by the lexer on an invalid character sequence; carries source position. */
public final class LexError extends RuntimeException {
    public final int line;
    public final int col;

    public LexError(String msg, int line, int col) {
        super("lexical error at " + line + ":" + col + ": " + msg);
        this.line = line;
        this.col  = col;
    }
}
