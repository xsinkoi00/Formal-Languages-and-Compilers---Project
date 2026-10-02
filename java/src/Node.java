/**
 * Common supertype for AST nodes. Two concrete kinds exist: {@link Num} and
 * {@link Op}. The checker uses {@code instanceof} to dispatch; keeping the
 * hierarchy shallow matches the Python implementation one-to-one.
 */
public abstract class Node {
    // Marker only — no shared state.
}
