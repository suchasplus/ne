# Fuzzy Search Design Document

This document outlines the design, algorithm selection, performance considerations, and final architectural decisions for the fuzzy search feature in the `ne` dictionary tool.

## 1. Feature Goal

When a user's search term does not yield an exact match in the database, the system should automatically perform a fuzzy search to find and suggest the most likely correct word based on spelling similarity. This handles common typos and improves user experience.

## 2. Core Algorithm: Levenshtein Distance

The core of the fuzzy search capability is the **Levenshtein distance** algorithm.

-   **Definition**: The Levenshtein distance between two strings is the minimum number of single-character edits (insertions, deletions, or substitutions) required to change one word into the other.
-   **Why it was chosen**: This algorithm is the industry standard for "edit distance" and perfectly models common user typing errors.
-   **Implementation**: To ensure correctness and performance, a well-established Go library, `github.com/agnivade/levenshtein`, was chosen for the implementation.

## 3. Evolution of Search Strategies & Selection

To efficiently implement fuzzy search on a static database of 600,000 entries, three tiers of solutions were evaluated.

### Tier 1: Optimized Linear Scan (Implemented Short-Term Solution)

This approach enhances performance at the application layer without altering the underlying KV database structure.

-   **Core Mechanism**: It uses a BoltDB `Cursor` to iterate through all keys, but intelligently avoids expensive distance calculations for every key.
-   **Key Optimizations**:
    1.  **Length Pruning**: Before calculating the distance, it first checks the length difference between the query term and the database word. If the difference is greater than the current best distance, the word is skipped. This is a low-cost, high-impact filter.
    2.  **Dynamic Threshold Adjustment**: The search begins with a preset maximum distance (e.g., `2`). If a match with a distance of `1` is found, the maximum distance for all subsequent comparisons is updated to `1`, aggressively pruning less relevant candidates.
-   **Pros**: Simple to implement, no changes to the DB structure, no extra storage overhead.
-   **Cons**: Still requires a physical iteration over all keys, so performance has a ceiling in the worst-case scenario.
-   **Conclusion**: Adopted and implemented as the short-term solution because it delivers the core feature quickly and its performance is acceptable for common typos of 1-2 characters.

### Tier 2: BK-Tree (Theoretically Optimal for Pure Distance Search)

This is an advanced solution specifically designed for nearest-neighbor searches.

-   **Data Structure**: A **BK-Tree** is a metric tree where nodes (words) are partitioned based on their Levenshtein distance to a parent node.
-   **Query Principle**: It leverages the **Triangle Inequality** property of metric spaces to prune entire branches of the tree, drastically reducing the number of comparisons.
-   **Pros**:
    -   **Ultimate Query Performance**: It is the theoretically optimal solution for pure similarity search, with a complexity approaching `O(log N)`.
-   **Cons**:
    -   **Low Space Efficiency**: Each node stores a full word, leading to a very large index size (estimated at >150MB for 600k words).
    -   **Extremely Low Versatility**: It **only** excels at similarity search and cannot efficiently support other query types like prefix search (autocomplete).

### Tier 3: Trie and its optimization, Radix Tree (The Versatile, High-Performance Solution)

This solution offers a balance of high performance, space efficiency, and functional versatility.

-   **Data Structure**: A standard **Trie** stores strings by sharing common prefixes. In practice, we use its optimized variant, the **Radix Tree** (or Patricia Trie), which compresses nodes with only one child into a single edge, significantly reducing the number of nodes.
-   **Query Principle**: Fuzzy search on a Trie/Radix Tree is performed with a specialized recursive algorithm that simulates the four edit operations (match, substitute, insert, delete) while traversing the tree, pruning branches when the accumulated distance exceeds the maximum allowed.
-   **Pros**:
    -   **High Space Efficiency**: By compressing paths, a Radix Tree's node count is dramatically reduced. Its final serialized size is comparable to a BK-Tree (estimated at 80-160MB) and potentially smaller.
    -   **Powerful Versatility**: **This is the decisive advantage of a Trie/Radix Tree.** A single data structure can efficiently power:
        1.  **Exact Match**
        2.  **Prefix Search** (for autocomplete)
        3.  **Fuzzy Search** (for spell correction)
-   **Cons**:
    -   **More Complex Fuzzy Search Algorithm**: The recursive implementation is more complex than the query algorithm for a BK-Tree.

## 4. Final Architectural Decision

**The Core Question: When Trie and BK-Tree have a similar storage cost, how do we choose?**

The answer lies in **opportunity cost** and **future architectural value**.

1.  **Resource Cost**: Both advanced solutions require a significant, and roughly equivalent, investment in storage and memory (~150MB).
2.  **Feature Return on Investment**:
    -   For ~150MB, a **BK-Tree** buys you one feature: **ultimate-speed fuzzy search**.
    -   For ~150MB, a **Radix Tree** buys you three features: **excellent fuzzy search**, **ultimate-speed prefix search**, and **ultimate-speed exact match**.

**Conclusion: The Radix Tree is the superior engineering choice.**

It provides a far greater return on investment in terms of features for a similar resource cost. Choosing a Radix Tree is a more forward-looking architectural decision, as it elegantly supports current and future needs with a single, unified data structure, avoiding future refactoring or architectural debt.

## 5. Final Decision: Phased Implementation

Based on this analysis, a phased implementation was decided:

1.  **Phase 1 (Implemented)**: Deliver the core feature quickly using the **Optimized Linear Scan (Tier 1)**, which is performant enough for most common use cases.
2.  **Phase 2 (Future Work)**: When the pursuit of ultimate performance becomes necessary, the project should **prioritize implementing the Radix Tree (Tier 3)**. The `kvbuilder` would be updated to build and serialize the Radix Tree, and `ne` would load it to power all query operations.
