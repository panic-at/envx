# envx — quick tour driven by Charm VHS (https://github.com/charmbracelet/vhs).
#
# Render the GIF:
#
#   go build -o /tmp/envxrec/envx ./cmd/envx
#   PATH=/tmp/envxrec:$PATH vhs docs/demo.tape
#
# Install VHS with `brew install vhs` or see the upstream README for other
# platforms.

Output docs/demo.gif

Set Shell         "bash"
Set FontSize      14
Set Width         1100
Set Height        680
Set Theme         "Catppuccin Mocha"
Set TypingSpeed   80ms
Set PlaybackSpeed 1.0

# Off-camera: drop into a clean working directory and clear the prompt noise.
Hide
Type "cd $(mktemp -d) && unset NO_COLOR PORT API_KEY && PS1='\[\e[1;36m\]\$\[\e[0m\] ' && clear"
Enter
Sleep 500ms
Show

Type "# envx — env vars across named profiles, injected into a process (never your shell)"
Sleep 600ms
Enter
Sleep 1500ms

Type "envx init"
Sleep 400ms
Enter
Sleep 1500ms

Type "envx profile add dev"
Sleep 400ms
Enter
Sleep 1500ms

Type "envx set PORT 8080 --profile dev"
Sleep 400ms
Enter
Sleep 1500ms

Type "envx set API_KEY demo-key-not-real --profile dev --sensitive"
Sleep 400ms
Enter
Sleep 1500ms

Type "envx show dev"
Sleep 400ms
Enter
Sleep 2s

Type "envx profile add prod --extends dev"
Sleep 400ms
Enter
Sleep 1500ms

Type "envx set PORT 443 --profile prod"
Sleep 400ms
Enter
Sleep 1500ms

Type "envx diff dev prod"
Sleep 400ms
Enter
Sleep 2s

Type "# Here is the point: envx run injects the vars into the child — and nowhere else."
Sleep 600ms
Enter
Sleep 1500ms

Type "echo PORT in this shell is: ${PORT:-EMPTY}"
Sleep 400ms
Enter
Sleep 1500ms

Type "envx run --profile dev -- sh -c 'echo PORT=$PORT API_KEY=$API_KEY'"
Sleep 400ms
Enter
Sleep 2s

Type "echo PORT in this shell is still: ${PORT:-EMPTY}"
Sleep 400ms
Enter
Sleep 2500ms
