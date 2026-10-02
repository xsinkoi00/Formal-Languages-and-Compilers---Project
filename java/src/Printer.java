import java.util.HashSet;
import java.util.Set;

/**
 * Infix pretty-printer with the minimum number of parentheses.
 *
 * For every operator child of an operator parent:
 *
 *   let pp = parent.prec, pc = child.prec
 *   wrap the child in '(' ... ')' iff
 *       pc <  pp,  OR
 *       pc == pp AND the child sits on the "wrong" side for the
 *                    parent's associativity:
 *                      - parent is right-assoc -> the LEFT child needs parens
 *                      - parent is left-assoc  -> the RIGHT child needs parens
 *
 * A {@link Num} child never needs parentheses. For the ternary (symbol
 * "?:") the middle operand sits between '?' and ':' and is syntactically
 * unambiguous regardless of its own precedence, so it is never wrapped.
 *
 * Spacing — chosen to reproduce all four assignment examples byte-for-byte:
 *     '+' '-' '*' '/' '%'   -> surrounded by single spaces
 *     '^' and '?:'          -> no spaces
 */
public final class Printer {

    private Printer() {}   // static utility

    private static final Set<String> SPACED = new HashSet<>();
    static {
        SPACED.add("+"); SPACED.add("-"); SPACED.add("*");
        SPACED.add("/"); SPACED.add("%");
    }

    // -----------------------------------------------------------------
    // Pretty-print
    // -----------------------------------------------------------------

    public static String pretty(Node node) {
        if (node instanceof Num) {
            return ((Num) node).text;
        }
        Op op = (Op) node;

        if ("?:".equals(op.symbol)) {
            String cond  = renderChild(op, op.children.get(0), Side.LEFT);
            String then_ = pretty(op.children.get(1));                    // middle is always safe
            String else_ = renderChild(op, op.children.get(2), Side.RIGHT);
            return cond + "?" + then_ + ":" + else_;
        }

        // Binary operator.
        String left  = renderChild(op, op.children.get(0), Side.LEFT);
        String right = renderChild(op, op.children.get(1), Side.RIGHT);
        String sep   = SPACED.contains(op.symbol) ? (" " + op.symbol + " ") : op.symbol;
        return left + sep + right;
    }

    private enum Side { LEFT, RIGHT }

    private static String renderChild(Op parent, Node child, Side side) {
        String rendered = pretty(child);
        if (child instanceof Num) return rendered;

        Op childOp = (Op) child;
        boolean needsParens = false;
        if (childOp.prec < parent.prec) {
            needsParens = true;
        } else if (childOp.prec == parent.prec) {
            // Same precedence: parens only on the side that fights the
            // parent's associativity.
            if (parent.rightAssoc && side == Side.LEFT)       needsParens = true;
            else if (!parent.rightAssoc && side == Side.RIGHT) needsParens = true;
        }
        return needsParens ? ("(" + rendered + ")") : rendered;
    }

    // -----------------------------------------------------------------
    // Human-readable AST dump (used by the --ast debug flag)
    // -----------------------------------------------------------------

    public static String astRepr(Node node, int indent) {
        StringBuilder sb = new StringBuilder();
        for (int k = 0; k < indent; k++) sb.append("  ");
        if (node instanceof Num) {
            sb.append("Num('").append(((Num) node).text).append("')");
            return sb.toString();
        }
        Op op = (Op) node;
        String assoc = op.rightAssoc ? "R" : "L";
        sb.append("Op('").append(op.symbol).append("', prec=").append(op.prec)
          .append(", ").append(assoc).append(")");
        for (Node c : op.children) {
            sb.append("\n").append(astRepr(c, indent + 1));
        }
        return sb.toString();
    }
}
