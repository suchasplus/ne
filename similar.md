# Fuzzy Search Design Document

This document outlines the design and implementation strategy for the fuzzy search feature in the `ne` dictionary tool.

## 1. Feature Goal

When a user's search term does not yield an exact match in the database, the system should automatically perform a fuzzy search to find and suggest the most likely correct word based on spelling similarity. This handles common typos and improves user experience.

## 2. Core Algorithm: Levenshtein Distance

The core of the fuzzy search capability is the **Levenshtein distance** algorithm.

-   **Definition**: The Levenshtein distance between two strings is the minimum number of single-character edits (insertions, deletions, or substitutions) required to change one word into the other.
-   **Why it was chosen**: This algorithm is the industry standard for "edit distance" and perfectly models common user typing errors, such as:
    -   **Insertion**: `devlop` → `develop` (1 edit)
    -   **Substitution**: `devrlop` → `develop` (1 edit)
    -   **Deletion**: `develo` → `develop` (1 edit)
-   **Implementation**: To ensure correctness and performance, a well-established Go library, `github.com/agnivade/levenshtein`, was chosen for the implementation.

## 3. Implementation Strategy & Efficiency Optimizations

A naive approach of iterating through hundreds of thousands of dictionary entries and calculating the Levenshtein distance for each would be too slow. The following optimizations were implemented in the short-term solution to ensure acceptable performance.

### 3.1. Length Pruning

-   **Logic**: This is a highly effective pre-filtering step. If the difference in length between two strings is greater than the maximum allowed Levenshtein distance, their actual Levenshtein distance *must* also be greater.
-   **Example**: If we are searching for a word similar to `devlop` (length 6) with a max distance of `2`, any word with a length less than `4` (6-2) or greater than `8` (6+2) can be safely skipped without performing the expensive distance calculation.
-   **Effect**: Dramatically reduces the number of candidates that need to be fully evaluated.

### 3.2. Dynamic Threshold Adjustment

-   **Logic**: This optimization focuses the search on finding the *best possible* match as quickly as possible. The search starts with a predefined maximum distance (e.g., `2`). If a match with a distance of `1` is found, `1` becomes the new maximum distance for all subsequent comparisons.
-   **Example**: When searching for `aply` with a max distance of `2`, the algorithm might first find `apple` (distance 2). It then continues, later finding `apply` (distance 1). At this point, the `bestDistance` is updated to `1`, and the suggestion list is reset to just `["apply"]`. Any future candidate with a distance greater than `1` will be ignored.
-   **Effect**: This ensures that once a very close match is found, the algorithm doesn't waste time evaluating less-likely candidates, leading to a faster conclusion.

## 4. Long-Term High-Performance Solution: BK-Tree

For ultimate performance, especially if the current solution proves insufficient, a more advanced data structure is recommended.

-   **What it is**: A **BK-Tree (Burkhard-Keller Tree)** is a specialized tree structure designed for similarity searching in metric spaces. Each node is a word, and its children are partitioned based on their Levenshtein distance to the parent node.
-   **How it works**: When searching, the tree is traversed by calculating the distance `d` between the query term and the current node. It then intelligently prunes entire branches of the tree by only exploring child nodes whose distance `d'` from the parent falls within the range `[d - maxDistance, d + maxDistance]`. This is based on the triangle inequality property and avoids a full database scan.
-   **Estimated Size**: For the `ne` project's dictionary of ~600,000 words, a serialized BK-Tree index was estimated to be between **100-200 MB**.
-   **Trade-offs**:
    -   **Pros**: Provides near-instantaneous fuzzy search results (`O(log N)` complexity).
    -   **Cons**: Significantly increases the database file size and the memory required by the `ne` tool during a fuzzy search, as the entire tree must be loaded into memory.

## 5. Final Decision: Phased Approach

Based on the analysis, a two-phased approach was decided upon:

1.  **Phase 1 (Implemented)**: Deliver a robust and reasonably fast solution using the **Levenshtein distance with Length Pruning and Dynamic Threshold Adjustment**. This provides the feature quickly without a massive increase in project complexity or resource requirements.
2.  **Phase 2 (Future Work)**: If the performance of the Phase 1 solution is found to be inadequate in real-world use, implement the **BK-Tree** solution. A potential optimization for this phase would be to only index the most frequently used subset of words to keep the index size and memory footprint manageable.
