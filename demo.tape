# Envforge Demo Tape

Output demo.gif

Set FontSize 14
Set Width 900
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 20

# Start clean
Type "clear"
Enter
Sleep 500ms

# 1. envforge info .env
Type "envforge info .env"
Enter
Sleep 2s

# 2. envforge diff
Type "envforge diff .env .env.example"
Enter
Sleep 2s

# 3. check
Type "envforge check --required PORT --from .env"
Enter
Sleep 2s

# 4. audit
Type "envforge audit ./testproject --env-file .env.example"
Enter
Sleep 3s

# 5. keygen
Type "envforge keygen"
Enter
Sleep 1s

# 6. version
Type "envforge version"
Enter
Sleep 2s

# End
Type "clear"
Enter
Type "echo Demo complete!"
Enter