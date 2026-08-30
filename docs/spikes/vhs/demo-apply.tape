# Demo 2 — the office switch: preview, apply, verify every frame, exit 0.
Output demo-apply.gif

Set Width 1080
Set Height 460
Set FontSize 14
Set TypingSpeed 15ms
Set Padding 14
Set WindowBar Colorful
Set BorderRadius 8

Hide
Type "export PATH=$HOME/.work/bin:$PATH && clear"
Enter
Show

Type `screenz apply --dry-run \`
Enter
Type `  --match 'bundle=com.microsoft.VSCode title=/office-demo/' --display 1 --region left-half \`
Enter
Type `  --match 'app="Google Chrome" title=/Example Domain/'          --display 1 --region right-half`
Sleep 500ms
Enter
Sleep 3.5s

Type `screenz apply \`
Enter
Type `  --match 'bundle=com.microsoft.VSCode title=/office-demo/' --display 1 --region left-half \`
Enter
Type `  --match 'app="Google Chrome" title=/Example Domain/'          --display 1 --region right-half`
Sleep 500ms
Enter
Sleep 3.5s

Type "echo exit=$?"
Sleep 300ms
Enter
Sleep 3s
