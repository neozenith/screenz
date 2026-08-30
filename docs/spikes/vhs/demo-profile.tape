# Demo 3 — save the context: profile save, commented YAML, replay.
Output demo-profile.gif

Set Width 1200
Set Height 470
Set FontSize 14
Set Padding 16

Hide
Type "export PATH=$HOME/.work/bin:$PATH && export SCREENZ_HOME=$(mktemp -d) && clear"
Enter
Show

Type "screenz profile save demo \"
Enter
Type "  --match 'bundle=com.apple.TextEdit title=/Untitled [0-9]/' \"
Enter
Type "  --display 2 --region grid=2x1 --gap 12"
Sleep 500ms
Enter
Sleep 2.5s

Type "cat $SCREENZ_HOME/profiles/demo.yaml"
Sleep 500ms
Enter
Sleep 4s

Type "screenz apply --dry-run demo"
Sleep 500ms
Enter
Sleep 4.5s
