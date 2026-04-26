# vhs tape for the README demo GIF.
# Render with: vhs demo/demo.tape  (writes demo/demo.gif)
# Install vhs: brew install vhs

Output demo/demo.gif

Set FontSize 14
Set Width 1100
Set Height 620
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 35ms

# Show the raw messy text — this is what you'd see and want to copy.
Type "cat demo/messy.txt"
Enter
Sleep 2s

# Copy it to the clipboard (simulates selecting the text and hitting Cmd+C).
Type "cat demo/messy.txt | pbcopy"
Enter
Sleep 1s

# Run ai-clean — reads clipboard, strips chrome, writes cleaned text back.
Type "ai-clean"
Enter
Sleep 2s

# Paste to reveal the cleaned output.
Type "pbpaste"
Enter
Sleep 3s
