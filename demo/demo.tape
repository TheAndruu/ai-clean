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

# Show the messy input as it arrived on the clipboard.
Type "cat demo/messy.txt"
Enter
Sleep 2s

# Run ai-clean on it via stdin so the cleaned output is visible inline.
Type "cat demo/messy.txt | ai-clean --stdin"
Enter
Sleep 4s
