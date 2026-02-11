## 2024-05-22 - TUI Loading State Layout Stability
**Learning:** In Terminal UIs (Bubble Tea), transient elements like loading spinners can cause jarring layout shifts if space isn't reserved.
**Action:** When adding a spinner that appears/disappears, reserve the vertical space (e.g., subtract an extra line from viewport calculation) even when it's hidden, or overlay it if possible. In this case, reserving a line `height - textarea - 2` kept the layout stable.

## 2024-05-23 - Error Visibility in TUI
**Learning:** Storing errors in the Bubble Tea model without explicit rendering logic (like appending to a history buffer) leaves users unaware of failures.
**Action:** Always implement a visual feedback mechanism for `errMsg` in `Update`, such as appending to the chat log with a distinct style.
