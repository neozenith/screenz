# Demo 2 — the office switch: preview, apply, verify every frame, exit 0.
Output demo-apply.gif

Set Width 1500
Set Height 500
Set FontSize 13
Set Padding 14
Set WindowBar Colorful
Set BorderRadius 8

Hide
Type "export PATH=$HOME/.work/bin:$PATH && clear"
Enter
Show

Type `screenz apply --dry-run \`
Enter
Type `  --match 'bundle=com.microsoft.VSCode title=/office-demo/' --display 1 --region maximize \`
Enter
Type `  --match 'app="Google Chrome" title=/neozenith/'          --display 2 --region left-half`
Sleep 500ms
Enter
Sleep 3.5s

Type `screenz apply \`
Enter
Type `  --match 'bundle=com.microsoft.VSCode title=/office-demo/' --display 1 --region maximize \`
Enter
Type `  --match 'app="Google Chrome" title=/neozenith/'          --display 2 --region left-half`
Sleep 500ms
Enter
Sleep 3.5s

Type "echo exit=$?"
Sleep 300ms
Enter
Sleep 3s
