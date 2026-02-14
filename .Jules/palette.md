## 2024-05-22 - TUI Loading State Layout Stability
**Learning:** In Terminal UIs (Bubble Tea), transient elements like loading spinners can cause jarring layout shifts if space isn't reserved.
**Action:** When adding a spinner that appears/disappears, reserve the vertical space (e.g., subtract an extra line from viewport calculation) even when it's hidden, or overlay it if possible. In this case, reserving a line `height - textarea - 2` kept the layout stable.

## 2024-05-24 - Standard Keybindings in TUI
**Learning:** Terminal users heavily rely on `Ctrl+L` to clear the screen, a reflex from shell usage. Missing this in a text-heavy TUI feels broken.
**Action:** Always implement `Ctrl+L` for clearing history/viewport in chat-based TUIs, and advertise it in the help footer for discoverability.
