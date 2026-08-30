# Demo 3 — save the office context: profile save, commented YAML, replay.
Output demo-profile.gif

Set Width 1360
Set Height 680
Set FontSize 13
Set Padding 14
Set WindowBar Colorful
Set BorderRadius 8

Hide
Type "export PATH=$HOME/.work/bin:$PATH && export SCREENZ_HOME=$(mktemp -d) && clear"
Enter
Show

Type `screenz profile save office \`
Enter
Type `  --match 'bundle=com.microsoft.VSCode title=/office-demo/' --display 1 --region maximize \`
Enter
Type `  --match 'app="Google Chrome" title=/neozenith/'          --display 2 --region left-half`
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
