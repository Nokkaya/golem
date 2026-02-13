## 2024-05-22 - TUI Loading State Layout Stability
**Learning:** In Terminal UIs (Bubble Tea), transient elements like loading spinners can cause jarring layout shifts if space isn't reserved.
**Action:** When adding a spinner that appears/disappears, reserve the vertical space (e.g., subtract an extra line from viewport calculation) even when it's hidden, or overlay it if possible. In this case, reserving a line `height - textarea - 2` kept the layout stable.

## 2024-05-24 - TUI Clear Screen Implementation
**Learning:** In Bubble Tea apps, implementing a clear screen (Ctrl+L) requires explicitly resetting the history buffer and refreshing the viewport content to avoid a "blank void" state.
**Action:** When adding clear screen functionality, always re-seed the history with a welcome message or context to maintain user orientation.
