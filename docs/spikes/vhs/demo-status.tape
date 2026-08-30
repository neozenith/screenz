# Demo 1 — trust, then see the world: doctor + scoped multi-match status.
Output demo-status.gif

Set Width 1360
Set Height 560
Set FontSize 13
Set Padding 14
Set WindowBar Colorful
Set BorderRadius 8

Hide
Type "export PATH=$HOME/.work/bin:$PATH && clear"
Enter
Show

Type "screenz doctor"
Sleep 500ms
Enter
Sleep 3.5s

Type `screenz status --match 'bundle=com.microsoft.VSCode title=/office-demo/' --match 'app="Google Chrome" title=/neozenith/'`
Sleep 500ms
Enter
Sleep 5s
