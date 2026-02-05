## 2024-05-22 - TUI Loading State Layout Stability
**Learning:** In Terminal UIs (Bubble Tea), transient elements like loading spinners can cause jarring layout shifts if space isn't reserved.
**Action:** When adding a spinner that appears/disappears, reserve the vertical space (e.g., subtract an extra line from viewport calculation) even when it's hidden, or overlay it if possible. In this case, reserving a line `height - textarea - 2` kept the layout stable.
