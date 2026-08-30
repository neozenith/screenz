# Demo 2 — the context switch: preview, apply, verify, exit 0.
Output demo-apply.gif

Set Width 1200
Set Height 380
Set FontSize 14
Set Padding 16

Hide
Type "export PATH=$HOME/.work/bin:$PATH && clear"
Enter
Show

Type "screenz apply --dry-run \"
Enter
Type "  --match 'bundle=com.apple.TextEdit title=/Untitled [0-9]/' \"
Enter
Type "  --display 2 --region grid=2x1 --gap 12"
Sleep 500ms
Enter
Sleep 3.5s

Type "screenz apply \"
Enter
Type "  --match 'bundle=com.apple.TextEdit title=/Untitled [0-9]/' \"
Enter
Type "  --display 2 --region grid=2x1 --gap 12"
Sleep 500ms
Enter
Sleep 3.5s

Type "echo exit=$?"
Sleep 300ms
Enter
Sleep 3s
