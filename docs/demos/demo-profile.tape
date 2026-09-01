# Demo 3 — save the office context: --save-profile, commented YAML, replay.
Output demo-profile.gif

Set Width 1120
Set Height 700
Set FontSize 14
Set TypingSpeed 15ms
Set Padding 14
Set WindowBar Colorful
Set BorderRadius 8

Hide
Type "export PATH=$HOME/.work/bin:$PATH && export SCREENZ_DEMO=demo-world.json && export SCREENZ_HOME=$(mktemp -d) && clear"
Enter
Show

Type `screenz apply --dry-run --save-profile office \`
Enter
Type `  --match 'bundle=com.microsoft.VSCode title=/office-demo/' --display built-in=true --region maximize \`
Enter
Type `  --match 'app="Google Chrome" title=/Example Domain/'          --display 1 --region left-half`
Sleep 500ms
Enter
Sleep 2.5s

Type "cat $SCREENZ_HOME/profiles/office.yaml"
Sleep 500ms
Enter
Sleep 4s

Type "screenz apply --dry-run --profile office"
Sleep 500ms
Enter
Sleep 4.5s
