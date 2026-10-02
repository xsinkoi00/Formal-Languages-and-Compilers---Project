import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.List;

/**
 * Expression Transformation Project — Java implementation.
 * Author: &lt;your full name, login&gt;
 *
 * Reads a prefix function-composition expression (e.g. "add(5, mul(3, 4))")
 * from standard input and prints the equivalent infix mathematical
 * expression (with the minimum number of parentheses) to standard output.
 *
 * Pipeline: stdin -&gt; Lexer -&gt; Parser -&gt; AST -&gt; Printer -&gt; stdout
 *
 * Debug entry points:
 *     java Expr --tokens &lt; input.txt    # dump the token stream
 *     java Expr --ast    &lt; input.txt    # dump the parsed AST
 *
 * Exit codes:
 *     0  success
 *     2  lexical error
 *     3  syntax error
 *     4  I/O error
 */
public final class Expr {

    public static void main(String[] args) {
        List<String> argv = Arrays.asList(args);

        String src;
        try {
            src = readAllStdin();
        } catch (IOException e) {
            System.err.println("I/O error: " + e.getMessage());
            System.exit(4);
            return;
        }

        try {
            if (argv.contains("--tokens")) {
                for (Token t : new Lexer(src).tokenize()) {
                    System.out.println(t);
                }
                return;
            }

            List<Token> tokens = new Lexer(src).tokenize();
            Node ast = new Parser(tokens).parse();

            if (argv.contains("--ast")) {
                System.out.println(Printer.astRepr(ast, 0));
                return;
            }

            System.out.println(Printer.pretty(ast));
        } catch (LexError e) {
            System.err.println(e.getMessage());
            System.exit(2);
        } catch (ParseError e) {
            System.err.println(e.getMessage());
            System.exit(3);
        }
    }

    /** Slurp stdin into a String. */
    private static String readAllStdin() throws IOException {
        BufferedReader r = new BufferedReader(
                new InputStreamReader(System.in, StandardCharsets.UTF_8));
        StringBuilder sb = new StringBuilder();
        char[] buf = new char[4096];
        int n;
        while ((n = r.read(buf)) != -1) sb.append(buf, 0, n);
        return sb.toString();
    }
}
