import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Recursive-descent parser over the prefix grammar:
 *
 *     expr ::= NUMBER
 *            | IDENT '(' expr (',' expr)+ ')'
 *
 * The input language is pure prefix, so operator precedence never enters the
 * parsing picture — a one-function recursive descent is enough. All
 * precedence/associativity work happens later, in {@link Printer}.
 */
public final class Parser {

    /** Metadata for one known function name. */
    private static final class FnInfo {
        final String  symbol;
        final int     prec;
        final boolean rightAssoc;
        final int     arity;
        FnInfo(String symbol, int prec, boolean rightAssoc, int arity) {
            this.symbol = symbol; this.prec = prec;
            this.rightAssoc = rightAssoc; this.arity = arity;
        }
    }

    /** Function name -> operator metadata. */
    static final Map<String, FnInfo> FUNCTIONS;
    static {
        Map<String, FnInfo> m = new HashMap<>();
        m.put("tern", new FnInfo("?:", 1, true,  3));
        m.put("add",  new FnInfo("+",  2, false, 2));
        m.put("sub",  new FnInfo("-",  2, false, 2));
        m.put("mul",  new FnInfo("*",  3, false, 2));
        m.put("div",  new FnInfo("/",  3, false, 2));
        m.put("mod",  new FnInfo("%",  3, false, 2));
        m.put("pow",  new FnInfo("^",  4, true,  2));
        FUNCTIONS = Collections.unmodifiableMap(m);
    }

    private final List<Token> toks;
    private int i = 0;

    public Parser(List<Token> tokens) { this.toks = tokens; }

    // -----------------------------------------------------------------
    // Entry point
    // -----------------------------------------------------------------

    public Node parse() {
        Node node = parseExpr();
        Token trailing = peek();
        if (trailing.kind != Token.Kind.EOF) {
            throw new ParseError(
                "unexpected trailing " + trailing.kind + " '" + trailing.lexeme + "'",
                trailing.line, trailing.col);
        }
        return node;
    }

    // -----------------------------------------------------------------
    // Cursor helpers
    // -----------------------------------------------------------------

    private Token peek() { return toks.get(i); }

    private Token advance() {
        Token t = toks.get(i);
        if (t.kind != Token.Kind.EOF) i++;
        return t;
    }

    private Token expect(Token.Kind kind, String what) {
        Token t = peek();
        if (t.kind != kind) {
            throw new ParseError(
                "expected " + what + ", got " + t.kind + " '" + t.lexeme + "'",
                t.line, t.col);
        }
        return advance();
    }

    // -----------------------------------------------------------------
    // Grammar rules
    // -----------------------------------------------------------------

    private Node parseExpr() {
        Token t = peek();

        if (t.kind == Token.Kind.NUMBER) {
            advance();
            return new Num(t.lexeme);
        }
        if (t.kind == Token.Kind.IDENT) {
            return parseCall();
        }
        throw new ParseError(
            "expected number or function name, got " + t.kind + " '" + t.lexeme + "'",
            t.line, t.col);
    }

    private Node parseCall() {
        Token name = advance();                     // IDENT
        FnInfo info = FUNCTIONS.get(name.lexeme);
        if (info == null) {
            throw new ParseError("unknown function '" + name.lexeme + "'",
                    name.line, name.col);
        }

        expect(Token.Kind.LPAREN, "'(' after function name");

        List<Node> args = new ArrayList<>();
        args.add(parseExpr());
        while (peek().kind == Token.Kind.COMMA) {
            advance();
            args.add(parseExpr());
        }

        expect(Token.Kind.RPAREN, "')' to close function call");

        if (args.size() != info.arity) {
            throw new ParseError(
                "'" + name.lexeme + "' expects " + info.arity +
                " argument(s), got " + args.size(),
                name.line, name.col);
        }

        return new Op(info.symbol, info.prec, info.rightAssoc, args);
    }
}
