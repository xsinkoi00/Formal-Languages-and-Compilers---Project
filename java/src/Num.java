/** Numeric literal leaf. The original source lexeme is kept verbatim. */
public final class Num extends Node {
    public final String text;
    public Num(String text) { this.text = text; }
}
