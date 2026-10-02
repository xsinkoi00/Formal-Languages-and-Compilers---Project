import java.util.List;

/**
 * Operator node. {@code symbol} is the infix spelling the printer will emit;
 * {@code prec} and {@code rightAssoc} drive the parenthesisation rule in
 * {@link Printer}. For the ternary we store three children and use the
 * composite symbol "?:".
 */
public final class Op extends Node {
    public final String      symbol;
    public final int         prec;         // 1 (?:) .. 4 (^)
    public final boolean     rightAssoc;
    public final List<Node>  children;

    public Op(String symbol, int prec, boolean rightAssoc, List<Node> children) {
        this.symbol     = symbol;
        this.prec       = prec;
        this.rightAssoc = rightAssoc;
        this.children   = children;
    }
}
