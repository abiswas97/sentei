Output "/tmp/sentei-vhs-progress-arc/outputs/before-after.gif"
Set Shell "bash"
Set Width 960
Set Height 672
Set FontSize 16
Set TypingSpeed 15ms
Set Framerate 30
Hide
Type "cd / && /Users/abiswas/code/personal/sentei/progress-arc/.demos/progress-arc/setup-fixture.sh >/dev/null && source /tmp/sentei-vhs-progress-arc/demo.env && cd /tmp/sentei-vhs-progress-arc/repo/demo.git && stty rows 24 cols 80" Enter
Show
Type "printf 'BEFORE  '; cat /tmp/sentei-vhs-progress-arc/bin/sentei-before.sha" Enter
Type "cd /tmp/sentei-vhs-progress-arc/repo/demo.git && /tmp/sentei-vhs-progress-arc/bin/sentei-before" Enter
Sleep 3s
Down 2
Enter
Sleep 3s
Type "a"
Sleep 1s
Enter
Sleep 1s
Type "y"
Sleep 5s
Type "q"
Sleep 1s
Type "q"
Sleep 1s
Hide
Type "cd / && /Users/abiswas/code/personal/sentei/progress-arc/.demos/progress-arc/setup-fixture.sh >/dev/null && source /tmp/sentei-vhs-progress-arc/demo.env && cd /tmp/sentei-vhs-progress-arc/repo/demo.git && stty rows 24 cols 80" Enter
Show
Type "printf 'AFTER   '; cat /tmp/sentei-vhs-progress-arc/bin/sentei-after.sha" Enter
Type "cd /tmp/sentei-vhs-progress-arc/repo/demo.git && /tmp/sentei-vhs-progress-arc/bin/sentei-after" Enter
Sleep 3s
Down 2
Enter
Sleep 3s
Type "a"
Sleep 1s
Enter
Sleep 1s
Type "y"
Sleep 5s
Type "q"
