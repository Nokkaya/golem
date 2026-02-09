## 2024-05-23 - Visual Error Feedback in TUI
**Learning:** TUI applications often fail silently or with subtle UI changes (spinner stopping) when backend errors occur. Explicitly rendering errors into the chat history ensures the user knows what happened and provides a persistent record of the failure.
**Action:** Always verify how `error` types are handled in `Update` loops of Bubble Tea applications. Ensure they are user-visible, not just internal state updates.
