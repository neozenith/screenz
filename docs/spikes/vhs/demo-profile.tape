# Demo 3 — save the office context: profile save, commented YAML, replay.
Output demo-profile.gif

Set Width 1080
Set Height 700
Set FontSize 14
Set TypingSpeed 15ms
Set Padding 14
Set WindowBar Colorful
Set BorderRadius 8

Hide
Type "export PATH=$HOME/.work/bin:$PATH && export SCREENZ_HOME=$(mktemp -d) && clear"
Enter
Show

Type `screenz profile save office \`
Enter
Type `  --match 'bundle=com.microsoft.VSCode title=/office-demo/' --display 1 --region left-half \`
Enter
Type `  --match 'app="Google Chrome" title=/Example Domain/'          --display 1 --region right-half`
Sleep 500ms
Enter
Sleep 2.5s

Type "cat $SCREENZ_HOME/profiles/office.yaml"
Sleep 500ms
Enter
Sleep 4s

Type "screenz apply --dry-run office"
Sleep 500ms
Enter
Sleep 4.5s
