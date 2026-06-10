# sentei fix-branch review — 2026-06-10

Independent adversarial review of this session's diff (c8e55d8..HEAD). Each finding verified against real behavior.

Totals: 18 confirmed | 10 refuted

## [P1] (migrate) Backup-phase failure prints a restore command that deletes the user's intact original repo

- **Where:** cmd/migrate.go / internal/repo/migrate.go:cmd/migrate.go:54-56; internal/repo/migrate.go:45,117,127
- **Scenario:** User runs `sentei migrate <repo>`. Validate passes, then the Backup phase's `cp -a` fails partway (e.g. disk full while doubling a large repo, or a permission error). The original root has NOT been touched yet (Migrate phase never ran).
- **Why wrong:** runMigrateBackup sets backupPath (migrate.go:117) BEFORE running cp, and on cp failure returns `phase, backupPath, ""` (line 127). Migrate() assigns result.BackupPath = backupPath (line 45) before the HasFailures() early-return. Both failure surfaces gate the restore message solely on `result.BackupPath != ""` (cmd/migrate.go:54 and migrate_summary.go:92) and print `To restore: rm -rf <BareRoot> && mv <BackupPath> <BareRoot>`. After a backup failure that command tells the user to rm -rf their fully intact repo and mv a partial/absent backup over it. The restore instruction is only valid for a Migrate-or-later failure, where the root was actually mutated.
- **Fix:** Only emit restore guidance when a valid restore source exists AND the root was modified: gate on the Backup phase having succeeded and the failing phase being Migrate or later (e.g. check findPhase(result.Phases,"Backup") succeeded before printing). Equivalently, do not set result.BackupPath when the Backup phase failed, or carry an explicit `BackupValid bool`.
- **Confidence:** high

## [P1] (protection) DetectDefaultBranch reads per-worktree HEAD: running `remove --all`/`--merged` from inside a worktree subdir unprotects (and can delete) the actual default-branch worktree

- **Where:** cmd/remove.go:48 (and 52)
- **Scenario:** Sentei bare repo whose default branch is non-standard, e.g. "production". User runs the non-interactive CLI with a repoPath that resolves inside a worktree subdir, e.g. `sentei remove --non-interactive --force --all ./myrepo/feature/x` (or cwd is the feature/x worktree and repoPath defaults to "."). DetectContext accepts the worktree subdir as ContextBareRepo (repo.go:68-76 via --git-common-dir ending in .bare), so RunRemove proceeds. git.DetectDefaultBranch(runner, repoPath) runs `git -C <worktree> symbolic-ref --short HEAD`, which returns the worktree's checked-out branch ("feature/x"), NOT the repo default ("production"). Protection then guards "feature/x" and treats "production" as an ordinary branch; `--all` removes the production worktree. Reproduced live: production worktree was REMOVED and feature/x was protected.
- **Why wrong:** HEAD is per-worktree in git, so `symbolic-ref --short HEAD` is cwd-sensitive. The pre-fix detector read repo-level refs (refs/remotes/origin/HEAD then `rev-parse --verify refs/heads/main`) and was cwd-independent, and pre-fix protection used convention-only IsProtectedBranch, so this path was previously safe. This fix unified detection onto the per-worktree HEAD and wired it into protection, introducing the regression. Both TUI entry points normalize first (main.go:308-310 runRoot and main.go:177 launchInteractiveDecision both call ResolveBareRoot before DetectDefaultBranch), but cmd.RunRemove does not -- it passes the raw repoPath. Same wrong defaultBranch also flows into CheckMerged: from inside a worktree that is ahead of the real default, the real default becomes an ancestor of the wrong target and is SELECTED by --merged for removal (reproduced: `--merged --dry-run` from inside feature/ahead listed "production" for removal). With --force this violates the safety-first tenet; the branch ref survives in .bare so it is recoverable, but any uncommitted/untracked changes in the wrongly-removed worktree are lost.
- **Fix:** In cmd.RunRemove, normalize repoPath to the bare root before detection, mirroring the TUI: after DetectContext confirms ContextBareRepo, set `repoPath = repo.ResolveBareRoot(runner, repoPath)` (verified safe for the existing root and .bare cases). Then DetectDefaultBranch, CheckMerged, ListWorktrees, and removal all operate against the bare root and HEAD resolves to the repo default. Add a regression test that runs ResolveFilters/RunRemove with repoPath pointing inside a worktree and asserts the real default is protected.
- **Confidence:** high

## [P1] (tui-summaries) Backup-phase failure renders data-destroying restore advice for an intact repo

- **Where:** internal/tui/migrate_summary.go:92-97
- **Scenario:** Migrate runs Validate (ok), then Backup, where `cp -a` fails (disk full, perms). Migrate() early-returns at migrate.go:47 BEFORE Phase 3, so the original .git is untouched and the repo is fully functional. The new `result.HasFailures()` check now treats this Backup failure as critical (previously only a 'Migrate' phase failure reached this screen). The summary then prints: 'Your original repo is backed up at: <path>' and restore command `rm -rf /tmp/my-project && mv /tmp/my-project_backup_... /tmp/my-project`.
- **Why wrong:** The backup copy FAILED, so BackupPath points at an absent/partial directory, yet the block renders because it is gated only on `BackupPath != ""`. result.BackupPath is non-empty: runMigrateBackup returns the constructed backupPath even on the failure branch (migrate.go:127), and Migrate assigns result.BackupPath at migrate.go:45 before the failure check at :47. A user who follows the advice runs `rm -rf` on a repo that was never broken, then `mv` fails (no source) -> total loss of a working repo. This is data-loss-grade advice on a plausible path.
- **Fix:** Make BackupPath reflect reality: in runMigrateBackup, return "" for backupPath on the copy-failure branch (migrate.go:127) instead of the constructed path. Then result.BackupPath stays empty on a backup failure -> the restore block does not render (matching the already-correct Validate-fail case), while a genuine Migrate-phase failure still has a real backup and still renders the restore block correctly. The identical sibling defect in cmd/migrate.go:54-57 (same `BackupPath != ""` gate, same RestoreCommand) is fixed by the same one-line change.
- **Confidence:** high

## [P1] (cli-dispatch) cleanup --force re-injection appended after positional repo path is silently dropped by the flag parser, so non-interactive force-delete is still broken (the fix is a no-op in its documented form)

- **Where:** main.go:146-153 (append at 151)
- **Scenario:** User runs the documented form `sentei cleanup --non-interactive --mode aggressive --force <repo>` (a positional repo path, exactly as the repo's own e2e tests cli_e2e_test.go:103 and :139 do), intending to force-delete branches whose upstream is gone and that are not fully merged.
- **Why wrong:** extractGlobalFlags strips every --force, so result.Args = [--mode aggressive <repo>]. main.go:151 does `args = append(args, "--force")` -> [--mode aggressive <repo> --force]. RunCleanup re-parses these with Go's flag package, which STOPS at the first non-flag token (<repo>) and does not resume, so --force is treated as a trailing operand and never sets the flag. opts.Force stays false, DeleteGoneBranches uses -d (not -D), and the unmerged branch is skipped with the message 'not fully merged — use --force' even though --force WAS supplied. The exact bug the re-injection was added to fix (audit finding flow-audit-2026-06-09.md:183) is therefore still live whenever a repo path is given.
- **Fix:** Prepend instead of append so the flag precedes any positional: `args = append([]string{"--force"}, result.Args...)`. Better still, avoid the string round-trip entirely: parse the cleanup flags once in main, set opts.Force = result.Force directly (mirroring the interactive path at main.go:198-199), and call RunCleanupWithOpts — eliminating the flag-ordering hazard altogether.
- **Confidence:** high

## [P2] (migrate) Copy phase writes through existing symlinks, corrupting the link target (incl. files outside the worktree)

- **Where:** internal/repo/migrate.go / internal/fileutil/copy.go:migrate.go:284-288 (copyDir/CopyFile); copy.go:9-19
- **Scenario:** The branch's checked-out tree (placed by `git worktree add`) contains a committed or untracked symlink; the backup contains the same symlink. The copy phase copies the backup entry over it. Particularly damaging when the symlink targets a file OUTSIDE the worktree (e.g. an untracked symlink to a shared dotfile).
- **Why wrong:** filepath.WalkDir/ReadDir do not follow symlinks, so a symlink entry has IsDir()==false and is routed to CopyFile, which does os.ReadFile(src) (follows the link) then os.WriteFile(dst, ...). Because git already checked out the committed symlink into the worktree, dst is an existing symlink; os.WriteFile writes THROUGH it to its target rather than replacing the link. I confirmed this clobbers the target file. For same-content committed symlinks it is a harmless no-op, but for differing content or a target outside the worktree it silently overwrites an external/unrelated file. Dangling symlinks also produce a scary 'no such file or directory' warning, and untracked symlinks are converted to regular files or dropped.
- **Fix:** Detect symlinks via entry.Type()&fs.ModeSymlink (and lstat inside copyDir's walk), and recreate them with os.Readlink + os.Symlink instead of reading/writing through them; remove an existing dst before writing. Skip/clearly report dangling links rather than emitting a raw open error.
- **Confidence:** high

## [P2] (create-creator) shellQuote fix not propagated to sibling {path} interpolation in integration/manager.go (TUI enable-integration flow still injectable)

- **Where:** internal/integration/manager.go:83
- **Scenario:** User has a worktree whose branch is a name git permits but the shell interprets, e.g. `feat$(touch /tmp/x)` or `feat;ls` (I confirmed git accepts all of these as valid branch names). They open the TUI integration list (or migrate-integrations) and enable an integration that has a Setup.Command containing {path}, e.g. crg's `code-review-graph build --repo {path}`. EnableIntegration iterates wtPaths (each is wt.Path, the branch-derived worktree dir) and runs `strings.ReplaceAll(integ.Setup.Command, "{path}", wtPath)` with NO quoting, then `shell.RunShell(workDir, cmd)` which is `sh -c`.
- **Why wrong:** This is the exact bug the session fixed in creator/integrations.go:171 via shellQuote, but the fix was applied to only one of the two consumers. manager.go:83 still does the raw ReplaceAll. The branch name flows unescaped into a `sh -c` string, so a legitimate branch like `feat$(date)` corrupts the command and an adversarial/accidental `feat$(rm -rf x)` executes. I reproduced it: with wtPath=`/repo/feat$(touch /tmp/SENTEI_MGR_INJ)` the unquoted path created /tmp/SENTEI_MGR_INJ and truncated the echoed path to `building /repo/feat`, while the fixed (quoted) creator path treated the same input as literal text and created no file. Same severity class as the creator-side injection the original audit fixed.
- **Fix:** Apply the same defense at manager.go:83: `cmd := strings.ReplaceAll(integ.Setup.Command, "{path}", shellQuote(wtPath))`. Better: hoist shellQuote into a shared location (e.g. internal/git or internal/integration) so both consumers reference one definition (DRY single-source) rather than duplicating the helper, since both interpolate the same {path} token into the same `sh -c` ShellRunner.
- **Confidence:** high

## [P2] (state-fileutil) RestoreCommand emits an unquoted destructive shell command; spaces in the repo path break it and rm -rf the wrong directories

- **Where:** internal/repo/migrate.go:367-369
- **Scenario:** User migrates a repo at a path with a space, e.g. ~/Documents/My Project/app. A later phase fails (e.g. clone --bare or worktree add), so the CLI (cmd/migrate.go:56) and TUI (internal/tui/migrate_summary.go:96) display RestoreCommand() as a copy-paste recovery command. The user pastes it to restore their original repo from the backup.
- **Why wrong:** RestoreCommand() builds `rm -rf %s && mv %s %s` with raw, unquoted paths. For BareRoot=/Users/x/My Project/app this produces `rm -rf /Users/x/My Project/app && mv ...`, which rm -rf interprets as two arguments: /Users/x/My and Project/app. The recovery command is destructive (rm -rf) and runs precisely when the migration already failed, so a copy-paste deletes unintended paths and the restore does not happen. The sibling backup step already does this correctly with `cp -a %q %q` (migrate.go:121), so the codebase knows these paths need quoting; RestoreCommand is the only place that omits it. Confirmed by running the exact function: input BareRoot="/Users/x/My Project/repo" yields `rm -rf /Users/x/My Project/repo && mv /Users/x/My Project/repo_backup_x /Users/x/My Project/repo`.
- **Fix:** Quote all three operands: `return fmt.Sprintf("rm -rf %q && mv %q %q", r.BareRoot, r.BackupPath, r.BareRoot)`. Add a table test with a spaced path asserting the operands are quoted.
- **Confidence:** high

## [P3] (clone) CLI clone event printer drops the new StepSkipped status — tracking-skip resolution is never shown

- **Where:** cmd/clone.go:58-71
- **Scenario:** A bare clone succeeds, but the subsequent best-effort `git fetch origin` or `--set-upstream-to` fails (e.g. a transient network/auth blip after the initial clone). runCloneWorktree emits Event{Step:"Set upstream tracking", Status:StepSkipped, Message:"no tracking: ..."}.
- **Why wrong:** printCloneEvent switches only on StepRunning/StepDone/StepFailed. The reorder introduced StepSkipped as a new status, but the CLI printer has no case for it, so the skip event prints nothing. The user sees `→ [Worktree] Set upstream tracking` (running) with no resolution line, then `✓ Cloned to ...`. The 'no tracking' explanation that the code carefully built into the step Message is silently discarded, so the user is never told their clone has no upstream configured.
- **Fix:** Add a `case repo.StepSkipped:` to printCloneEvent that prints the step with a skip glyph and e.Message (e.g. `⊘ [Phase] Step (message)`), mirroring the StepDone branch.
- **Confidence:** high

## [P3] (clone) TUI live progress counts a StepSkipped step as neither done nor failed, leaving the Worktree phase stuck at <100% during the hold window

- **Where:** internal/tui/repo_progress.go:127-135, 170
- **Scenario:** Same tracking-skip scenario as above, in the interactive TUI. The Worktree phase has 3 steps (Detect default branch=Done, Create worktree=Done, Set upstream tracking=Skipped).
- **Why wrong:** buildRepoPhaseDisplays only increments pd.done for StepDone and StepFailed; a StepSkipped step increments neither. So pd.done=2 while pd.total=3, making `isComplete := pd.done == pd.total` false. The phase header renders as active/incomplete at 66% even though the clone has finished successfully. With a non-zero minProgressDuration the progress view (showing the perpetually-incomplete Worktree phase) lingers for the remaining hold before holdOrAdvance switches to the summary, so the glitch is user-visible, not instantaneous. (The downstream summary view is correct — it reports 'ready'.)
- **Fix:** In buildRepoPhaseDisplays count StepSkipped toward pd.done (a skipped step is a completed, non-failing step), e.g. add `case creator.StepSkipped: pd.done++` to the switch at lines 128-134 so the phase reaches 100% when all steps are done-or-skipped.
- **Confidence:** high

## [P3] (clone) No test asserts the repo dir is preserved on a tracking-skip (the keep-on-skip filesystem effect is unverified)

- **Where:** internal/repo/clone_test.go:156-181
- **Scenario:** TestClone_FetchFailure_StillSucceedsWithoutTracking drives `fetch origin` to fail and asserts !result.HasFailures() and WorktreePath != "".
- **Why wrong:** The rollback decision at clone.go:106 is `if wtPhase.HasFailures() && !worktreeCreated`. On a tracking-skip the worktree WAS created, so the dir must be kept. The rollback-removes test (TestClone_WorktreeFailure_RollsBackPartialDir) verifies the destructive branch with a real os.Stat, but the symmetric keep branch is only verified via WorktreePath (an in-memory field), not by asserting the on-disk repoPath still exists. A future regression that made the skip path call rollback would leave HasFailures()/WorktreePath correct yet delete a usable checkout, and this test would still pass.
- **Fix:** Add an onRun hook to that test (as TestClone_WorktreeFailure_RollsBackPartialDir does) so the worktree add creates a real dir, then assert os.Stat(repoPath) succeeds after the tracking-skip — proving the usable checkout is preserved.
- **Confidence:** medium

## [P3] (migrate) No test covers RestoreCommand, the CLI/TUI restore-on-failure output, or the copy/backup failure paths

- **Where:** internal/repo/migrate_test.go:148-252 (new tests)
- **Scenario:** The three data-loss bugs above (A, B, C) all live in code that the new tests do not exercise.
- **Why wrong:** The new tests are mock-based and assert that the expected git commands were issued (somewhat tautological for the slash-branch/origin cases — they verify the code calls what the code calls, not git's actual behavior). There is no test for RestoreCommand() output, no test asserting BackupPath is empty/restore is suppressed when the Backup phase fails, and no test asserting the Copy phase fails (or blocks delete) when files cannot be copied. These are exactly the paths where the real bugs are.
- **Fix:** Add: (1) a RestoreCommand test including a spaces-in-path case; (2) a Migrate test with a failing ShellRunner asserting the Backup failure does not produce a usable destructive restore command; (3) a runMigrateCopy test asserting the phase reports failure when copies fail; consider one real-git e2e covering origin restore + worktree placement so the assertions test git behavior, not the mock.
- **Confidence:** high

## [P3] (tui-summaries) No TUI-layer tests for any migrate-summary path (scan-all-phases logic and restore block untested)

- **Where:** internal/tui/migrate_summary.go:23,67-104
- **Scenario:** The diff rewrote both updateMigrateSummary and viewMigrateSummary to scan ALL phases via result.HasFailures() instead of only the named 'Migrate' phase, and switched the restore line to result.RestoreCommand(). No migrate_summary_test.go was added or exists.
- **Why wrong:** The most behavior-changing edit in this area (which phases are 'critical', the failErr extraction via FirstFailure, the BackupPath-gated restore block, and the key-handling that locks the screen to q-only on failure) has zero unit coverage. By contrast the clone/create summaries got a thorough repo_summary_test.go. The untested restore-block path is exactly where the P1 above hides; a crafted-result view test would have caught it.
- **Fix:** Add internal/tui/migrate_summary_test.go with crafted MigrateResults: (a) Backup-phase failure asserts NO restore block / NO 'rm -rf' once the P1 is fixed; (b) Migrate-phase failure asserts the restore block IS shown with the real backup path; (c) Validate-phase failure asserts q-only and no restore block; (d) success path asserts 'migrated' + 'Delete backup?'. Drive viewMigrateSummary and updateMigrateSummary directly (the pattern in repo_summary_test.go).
- **Confidence:** high

## [P3] (tui-summaries) "GitHub" soft/hard discriminator duplicated as a control-flow magic string

- **Where:** internal/tui/repo_summary.go:29,118
- **Scenario:** The phase name 'GitHub' is the literal that distinguishes a soft (unpublished, relaunch-able) failure from a hard (broken local repo, quit) failure. It is hardcoded independently in createHardFailed (line 29: `phase.Name != "GitHub"`) and again in viewCreateRepoSummary (line 118: `phase.Name == "GitHub"`).
- **Why wrong:** Two copies of the same control-flow-driving string must stay in lockstep with create.go's Phase{Name: "GitHub"}. If the phase is ever renamed or a second soft-failure phase is added, a silent divergence makes createHardFailed and the render disagree (e.g. relaunch gate says soft, render says hard). Not a bug today since create.go only emits Setup+GitHub, but it violates single-source-of-truth for a value that gates whether the user is sent into a broken repo.
- **Fix:** Define the phase name once (e.g. a const phaseGitHub = "GitHub" in the repo package, alongside where create.go constructs Phase{Name: ...}) and reference it from both createHardFailed and viewCreateRepoSummary. Ideally expose the soft/hard classification as a method on CreateResult so the TUI does not re-implement it.
- **Confidence:** high

## [P3] (tui-plumbing) integrationFinalizedMsg save-error guard (err != nil) is untested

- **Where:** internal/tui/integration_progress.go:39
- **Scenario:** state.Save fails (disk full, permission denied, read-only .bare) during integration apply on the normal (non-migrate) path. The new guard `if msg.err == nil && m.integ.returnView != migrateNextView` is specifically designed to NOT mutate m.integ.current / m.integ.staged and NOT bump worktreeGeneration in that case, so the in-memory integration set stays consistent with what is actually on disk.
- **Why wrong:** The entire purpose of the fix is the err!=nil behavior, but every test that drives updateIntegrationProgress with integrationFinalizedMsg passes err: nil (integration_progress_test.go:56 and :85). menu_test.go:239 sends err on a worktreeContextMsg, not this message. So the regression-relevant branch (state must NOT be mutated when the save failed) has zero coverage; a future edit that drops the `msg.err == nil` clause would silently pass CI while reintroducing the bug of showing an integration set that was never persisted.
- **Fix:** Add a table case to TestUpdateIntegrationProgress_FinalizedMsg that sends integrationFinalizedMsg{current: newCurrent, err: errors.New("save failed")} with returnView=integrationListView, then assert m.integ.current/staged are UNCHANGED from their pre-message values and that the returned cmd does not include a loadWorktreeContext (i.e. worktreeGeneration did not advance). This pins the exact contract the guard added.
- **Confidence:** high

## [P3] (cli-dispatch) Empty-stderr %w fallback in GitRunner.Run / RunShell has no direct unit test despite available fake-git harness

- **Where:** internal/git/commands.go:25-31 (Run), 51-55 (RunShell)
- **Scenario:** The fix's core behavior — when git exits non-zero with empty stderr, fall back to wrapping the exec error with %w instead of producing a content-free 'git <args>: ' message — is the new code path. A future refactor could drop the %w branch (or swap it for the old %s) and every existing test would still pass.
- **Why wrong:** Only ValidateRepository's wrap is unit-tested (TestValidateRepository_PreservesUnderlyingCause), and it mocks the runner, so it never exercises the real GitRunner.Run stderr-empty branch. gitrunner_test.go already has installFakeGit + TestHelperProcess, which could add a case that exits non-zero with empty stderr and assert the message falls back to the exec error rather than a trailing-colon blank. The actual changed lines are untested.
- **Fix:** Add a TestHelperProcess case (e.g. a marker dir) that exits non-zero writing nothing to stderr, then assert GitRunner.Run returns a non-empty, cause-bearing error (errors.Unwrap non-nil / message contains 'exit status'). Optionally one for the non-empty-stderr branch to lock both halves.
- **Confidence:** high

## [P3] (cli-dispatch) No test sets up an unmerged gone-upstream branch to assert cleanup force-delete behavior, so both the re-injection ordering bug and the safe-mode footgun pass CI

- **Where:** cmd/cli_e2e_test.go:103, 139 (safe-mode cleanup tests) and cmd/cleanup_flags_test.go:34
- **Scenario:** The e2e tests TestCleanup... run `cleanup --mode safe --non-interactive --force <repo>` but only assert the command exits 0 / prints output. cleanup_flags_test.go:34 asserts ParseCleanupFlags maps a directly-supplied --force to Force=true, which is true in isolation but does not reflect what the dispatcher actually passes (it appends --force after the positional, where it is dropped).
- **Why wrong:** No test constructs a repo containing an unmerged branch with a gone upstream and asserts whether it is deleted vs skipped after a non-interactive cleanup. That single fixture would have caught both higher-severity findings: the positional-path no-op (branch wrongly survives in aggressive mode) and the safe-mode footgun (branch wrongly deleted in safe mode). The TestParseCleanupFlags_Force test gives false confidence because it bypasses the dispatch/append path where the bug lives.
- **Fix:** Add an e2e (or main-package dispatch) test with an unmerged gone-upstream branch covering the matrix: aggressive+force with a positional path must delete it (currently fails — would catch finding 1); safe+force must NOT delete it (currently deletes — would catch finding 2). Drive the real binary / Dispatch so the --force re-injection ordering is exercised, not ParseCleanupFlags directly.
- **Confidence:** high

## [P3] (create-creator) Push-failure orphan message is untested (no assertion on wrapped message or preserved error chain)

- **Where:** internal/repo/create.go:250
- **Scenario:** The push to GitHub fails (e.g. auth/network), so the code wraps the original error with `fmt.Errorf("%w (an empty GitHub repo %q now exists; ...)", err, opts.Name)`. No test exercises this branch.
- **Why wrong:** The fix's whole value is (a) the user-facing orphan-repo guidance string and (b) keeping the original error unwrappable via %w so callers/errors.Is still work. TestCreate_WithGitHub mocks a successful push, and TestCreate_GitHubPhaseFailure_LocalStillUsable fails earlier at user-lookup (api user), so neither reaches line 246-254. A future refactor could silently drop the %q name, the guidance, or downgrade %w to %v with no test catching it. (I separately verified by hand that the %w wrapping works and errors.Is finds the original, so the code is correct today; the gap is coverage.)
- **Fix:** Add a CreateWithGh test where the gh user-lookup, repo create, and `remote set-url` mocks succeed but the `main:[push -u origin main]` mock returns an error. Assert the github phase has a Push step that is StepFailed, that its Error message contains opts.Name and the orphan-repo guidance substring, and that errors.Is(step.Error, originalErr) is true.
- **Confidence:** high

## [P3] (state-fileutil) No test coverage for RemoveAllRetry, the Save fsync error/cleanup path, or RestoreCommand

- **Where:** internal/fileutil/remove.go:9
- **Scenario:** Future edits to RemoveAllRetry, the Save temp-file lifecycle, or RestoreCommand have no regression guard for their error/edge behavior.
- **Why wrong:** internal/fileutil has no remove_test.go at all. internal/state/state_test.go does not assert that Save flushes (Sync) or that the temp file is removed on the sync-error path. RestoreCommand has no test. The Spotlight race itself is genuinely not unit-testable (nondeterministic), so the realistic gap is happy-path + error-path coverage: RemoveAllRetry returns the last error on a path that cannot be removed and returns nil on a populated dir; Save leaves no .sentei-*.json temp file behind after a failure; RestoreCommand quotes its operands (which would have caught the P2 above).
- **Fix:** Add internal/fileutil/remove_test.go (success on a populated dir; last-error returned when the parent is read-only). Add a Save test asserting no leftover .sentei-* temp files after success and that the file is fsynced/readable. Add a RestoreCommand test with a spaced path asserting quoting.
- **Confidence:** high

---
## Refuted (not real / pre-existing — recorded)

- (clone) Rollback swallows RemoveAllRetry error, leaving a half-built dir with no notice on cleanup failure
  - The finding is technically accurate about the code (the error IS discarded and the comment does say "leave nothing half-built behind"), but it does not meet the bar for a confirmed bug introduced by the fix, on three independent grounds.

1) No consumer mishandles the case. The clone failure is reported entirely independently of rollback. cmd/clone.go:44-50 iterates phases and returns the failed step's StepFailed error; repo_summary.go cloneFailed() reads repo.FirstFailure(). The user is always told the clone failed and why, whether or not RemoveAllRetry succeeded. Neither consumer depends on the directory being gone, and neither has a retry loop that breaks on a leftover dir. There is no inconsistent or broken repo state: the worst case is a half-built dir on disk plus a correctly-reported clone failure, which is an incomplete courtesy-cleanup, not a corrupt state.

2) It is not a regression; it is a strict improvement. `git show c8e55d8:internal/repo/clone.go` confirms no rollback existed before this diff. Pre-diff, EVERY failed clone left the half-built dir behind unconditionally, with no notice. Post-diff, the dir is cleaned up in the normal case and, only in the rare event that RemoveAllRetry exhausts all 50 retries, left behind with no notice. That rare-case behavior is identical to the old baseline, so the fix cannot have introduced a regression. A retry by the user then hits validateCloneTarget's accurate "target already exists" message (also new, also actionable), and pre-diff a retry would likewise have failed.

3) Swallowing the error is the correct pattern here, not a bug. This is best-effort secondary cleanup that runs AFTER the primary failure has already been surfaced. Reporting the cleanup error would muddy the actionable error the user actually needs. This is distinct from migrate.go:168, which correctly surfaces its RemoveAllRetry error because there removing the original .git is a required primary step of the migration, not post-failure cleanup. The "leave nothing half-built behind" comment slightly overstates the best-effort guarantee, but a doc-wording nit is not a behavioral bug and does not meet the task's confirmation bar (a reproduced wrong behavior or a traced consumer mishandle). A live experiment was deliberately not run because it can only reproduce the benign, reported, non-regressive outcome (dir remains + failure reported) and cannot produce the consumer-mishandling the bar requires.
- (migrate) Copy phase never marks itself failed, so delete-backup destroys the only copy of dropped files
  - The reviewer's MECHANIC is real and I reproduced it: a file that fails to copy is emitted only as a StepRunning "could not copy" warning, runMigrateCopy still appends a StepDone step ("1 items restored"), phase.HasFailures()==false, and the failed file is absent from the worktree — so deleting the backup loses it. Both consumers (cmd/migrate.go:67 and internal/tui/migrate_summary.go:41-42) gate DeleteBackup on result.HasFailures(), which the copy phase can never trip.

BUT the task is to judge whether the FIX introduced this, and the git history refutes that:

1. Pre-fix runMigrateCopy (git show c8e55d8:internal/repo/migrate.go) ALREADY emitted per-file copy failures as StepRunning warnings ("warning: could not copy %s") and ALWAYS appended a StepDone step. The "copy phase never marks itself failed" pattern is byte-for-byte pre-existing — the diff did not introduce it.

2. Pre-fix cmd/migrate.go (git show c8e55d8:cmd/migrate.go) had the IDENTICAL `if opts.DeleteBackup && result.BackupPath != ""` block calling repo.DeleteBackup, and pre-fix internal/tui/migrate_summary.go had the IDENTICAL `_ = repo.DeleteBackup(result.BackupPath)` on the Yes key. Both delete paths and their HasFailures() gating existed before this diff.

So the full chain (copy fails → no StepFailed → backup deleted → data loss) existed verbatim before the diff. This is the explicit refutation ground in the task: pre-existing, not introduced.

3. The diff actually REDUCED the data-loss surface, not increased it. Pre-fix the copy phase copied only a fixed allowlist (.env*, node_modules, vendor, build, dist, .vscode, .idea), so deleting the backup silently lost ALL other untracked/ignored/modified files even on a fully successful copy. Post-fix it copies everything except .git/.bare, making the backup genuinely redundant in the common (no-error) case. The fix made --delete-backup safer.

The residual risk the reviewer reproduced is the unchanged pre-existing per-file mechanic. One honest nuance: the diff DID add a new comment ("makes the backup genuinely redundant and safe to delete") that the code does not guarantee on a partial copy failure — a minor comment over-claim, not a runtime regression. Note also --delete-backup defaults to false (cmd/migrate_flags.go:18), so the dangerous path requires explicit opt-in.

Verdict: refuted as a bug introduced by the fix. The mechanic is real but pre-existing and the diff narrowed, not widened, the loss window.
- (migrate) RestoreCommand() is unquoted and becomes destructive on paths with spaces
  - The finding's CORE CLAIM — that the FIX introduced a destructive unquoted restore command — is FALSE. The unquoted format is pre-existing, not introduced by this diff.

Evidence of pre-existence: At base commit c8e55d8, internal/tui/migrate_summary.go:102-104 already contained `fmt.Fprintf(&b, "    rm -rf %s && mv %s %s\n", result.BareRoot, result.BackupPath, result.BareRoot)` — byte-for-byte identical unquoted format. This diff did NOT touch quoting; it refactored that inline string into a shared method MigrateResult.RestoreCommand() (migrate.go:367-369) and pointed both the TUI and a NEW cmd/migrate.go consumer at it. The format string is unchanged. The reviewer misattributed a pre-existing defect to the fix.

The TECHNICAL defect itself is real, and I confirmed it with a live experiment. Under a path with a space, the printed literal `rm -rf /tmp/.../My Projects/repo && mv ...` word-splits: `rm -rf` received the first token `/tmp/.../My` and deleted an unrelated directory (my decoy `My/SHOULD_NOT_BE_DELETED.txt` was destroyed), and the subsequent `mv` failed with "Projects/repo is not a directory". Reachability confirmed: BareRoot = opts.RepoPath = filepath.Abs(".") (main.go:167-169), the user's cwd, which routinely contains spaces on macOS, with no escaping applied. The string is display-only (CLI stderr, TUI failure screen) for manual copy-paste — sentei does not execute it. The reviewer's point that the codebase knows quoting is required is valid: migrate.go:121 uses `cp -a %q %q` via shell.RunShell.

Why severity is adjusted from P1-introduced down to P2-preexisting: (1) Not introduced by this diff, so it cannot be a regression/fix-bug — the premise of this review pass. (2) The defect triggers only on an uncommon path (spaces in the repo path), only at recovery time after a migration failure, and only if the user blindly copy-pastes. (3) The refactor is actually a mild net positive: centralizing into RestoreCommand() means the proper fix (swap to `%q`) now touches one declaration instead of two. The one honest nuance: the diff did PROPAGATE the unquoted format to a new CLI consumer, marginally widening the pre-existing blast radius — but it created nothing new. Root-cause fix (for the underlying pre-existing bug, not a fix-regression): change RestoreCommand to `fmt.Sprintf("rm -rf %q && mv %q %q", r.BareRoot, r.BackupPath, r.BareRoot)`, matching the backup phase's existing %q convention. File: /Users/abiswas/code/personal/sentei/fix-detect-default-branch/internal/repo/migrate.go:368.
- (migrate) RemoveAllRetry retries permanent errors in a tight 50x busy-loop
  - REFUTED — the described behavior is real but it is not a bug.

The reviewer's literal description is accurate: remove.go:11 loops 50x, with no backoff and no error classification. But "retries permanent errors in a tight busy-loop" is not a defect here, on four grounds:

1. EMPIRICAL TIMING. Live experiment: a permanent EACCES failure runs all 50 iterations in ~1.81ms total and returns the correct wrapped error ("unlinkat ...: permission denied"). The success path returns in ~120µs (one iteration). The reviewer's own writeup concedes "Harmless on a small .git (fast)". There is no busy-wait of any consequence and the error is faithfully propagated, not swallowed.

2. NO MISHANDLING CONSUMER. All three callers handle the returned error correctly: migrate.go:168 marks the step StepFailed and returns; clone.go:76 uses it only in a best-effort rollback (error intentionally discarded with `_`); playground/setup.go:111 returns the error to its caller. No consumer leaves inconsistent state or ignores a real failure.

3. NEAR-UNREACHABLE AT THE FLAGGED CALL SITE. The reviewer's headline scenario ("read-only mount / permission denied") cannot realistically reach migrate.go:168. The preceding step (migrate.go:153) does `git clone --bare .git barePath` where barePath lives under the same parent as repoPath/.git. If that location were not writable, the bare clone would already have failed and returned at line 158 before RemoveAllRetry is ever called. A genuinely permanent, location-level write failure is therefore filtered out upstream.

4. THE IMPLIED FIX IS OVER-ENGINEERING. Adding ENOTEMPTY-vs-EACCES classification plus backoff/jitter would add real complexity for ~2ms of avoided spinning on an effectively unreachable path. That directly conflicts with the project's "No Over-Engineering" / "No Shortcuts→but no speculative complexity" rules. The function is a deliberate, documented design choice (the doc comment explains the macOS Spotlight/ENOTEMPTY rationale and the intentional "no sleep"); it is newly introduced by this diff but is neither a regression nor a logic error.

Honesty note: the experiment measured a single-file permanent error, not a large unremovable tree. A large tree would re-walk on each pass, but os.RemoveAll removes what it can per pass so subsequent passes shrink, the loop is still bounded at 50, and the correct error is still returned — and the reachability argument (point 3) makes even that case moot at the migrate call site.

Not introduced-broken, not a plausible-scenario break, not a real correctness/error-handling bug. Refuted; severity N/A.
- (protection) protectedCount under --merged undercounts: the default-branch worktree is not reported as a protected skip
  - The reviewer's raw observation is factually reproducible but its framing as a "bug" / "inconsistency" is wrong. The metric the code implements is filter-relative: "Skipped (protected): N" = protected worktrees the ACTIVE filter would have selected but protection blocked. The reviewer silently substitutes a different metric (absolute count of all protected worktrees) and then calls the code inconsistent for not matching it. My two consistency experiments dismantle that:

1. --all run (production default + develop + feature/x): "Removed: 1, Skipped (protected): 2" — BOTH production and develop counted, because matchesFilters returns true for opts.All. So the default branch IS counted when the filter targets it.

2. --stale 30d run (only feature/old is stale; develop and production fresh): "Removed: 1" with NO "Skipped (protected)" line at all (count=0). Neither protected branch is stale, so the --stale filter never wanted them, so protection didn't "skip" them. Correctly excluded.

3. --merged run (the reviewer's scenario): "Removed: 1 (feature/merged), Skipped (protected): 1 (develop only)". This is correct, not an undercount: I verified `git merge-base --is-ancestor develop production` succeeds, so develop genuinely is a --merged candidate that protection saved — rightly counted. Production is the merge target; CheckMerged (remove_filter.go:78-80) deliberately self-skips the default branch because a branch being its own ancestor is meaningless for "is this a removal candidate." So --merged never targets production; protection didn't need to save it from --merged; reporting it as a --merged skip would be noise.

The rule is therefore uniform across all three filters, not inconsistent. It is also strictly BETTER than the pre-diff behavior: c8e55d8 counted every protected non-bare worktree unconditionally (git.IsProtectedBranch with no matchesFilters gate), which under --stale here would have falsely printed "Skipped (protected): 2" for two fresh branches the filter never targeted. The matchesFilters gate is new in this diff, so if it were a bug it would be a newly-introduced one — but it isn't a bug, it's the documented intent (comment at remove.go:57-59) and an accuracy improvement.

Provenance: confirmed the gating is introduced by this diff (pre-diff counted unconditionally), so the "pre-existing?" branch resolves cleanly — new code, correct behavior. Deletion behavior is unaffected in every run (the reviewer concedes this), so even under a hostile reading it could never exceed P3; and since the behavior is correct-by-design with no residual defect, severity is N/A.

One out-of-scope note (not part of this finding, do not block on it): the protectedCount summary line is not directly unit-tested because it lives inside RunRemove's side-effecting orchestration. That is a general coverage observation, not a correctness defect in the --merged path this finding concerns.
- (tui-plumbing) Pre-existing (NOT introduced by this diff): interactive `remove --merged/--all/--stale` renders "No worktrees found"
  - The described behavior is REAL and reproduces: an in-memory test wiring the model exactly as main.go does for `remove --merged` (NewMenuModel(ContextBareRepo) -> SetRemoveOpts) showed Init() returns nil (the fake runner's Run was never called, so loadWorktreeContext never fires), m.remove.worktrees stays at 0, and View() renders "No worktrees found." with no [P] marker. "Refuted" here means "not a bug in THIS fix," NOT "behavior does not happen" — the behavior does happen.

It is refuted as a fix-bug because it is PRE-EXISTING, which the task lists as a refute condition. `git show c8e55d8:internal/tui/model.go` confirms both Init (gates `m.view == menuView`) and SetRemoveOpts (sets `m.view = listView`, only marks pre-selected paths among already-loaded worktrees) are byte-identical to base. The diff's only model.go change adds `defaultBranch` to removeState and populates it from worktreeContextMsg (model.go:442-443); it does not touch Init, SetRemoveOpts, or the loading gate. The main.go SetRemoveOpts interactive wiring also predates the diff (base main.go:229, HEAD main.go:248), changed only by unrelated defaultBranch threading into ResolveFilters.

The "incomplete fix" door is also closed structurally: worktrees and defaultBranch arrive in the SAME worktreeContextMsg and are set together (model.go:442-443). There is no listView code path that loads worktrees without also setting defaultBranch — they are coupled in one message. So on this empty-list path there is simply nothing to protect; the unset defaultBranch is genuinely moot, not a gap in the fix's threading. The defaultBranch fix correctly threads the value everywhere worktrees are actually populated (the menu flow via worktreeContextMsg).

Net: the behavior reproduces but is a pre-existing UX quirk byte-identical to base, not introduced, worsened, or left incomplete by this diff. Out of this diff's mandate.
- (cli-dispatch) Re-injected --force couples the mandatory destructive gate to cleanup.Options.Force, so non-interactive --mode safe force-deletes unmerged branches with no way to opt out
  - The finding's TECHNICAL FACTS all reproduce exactly, but its characterization as a "bug introduced by the fix" is wrong. The behavior is the faithful, sanctioned implementation of the audit's own Option B, and the reviewer imports a meaning of "safe mode" the tool never promised.

WHAT I CONFIRMED LIVE (mechanism is real):
- TEST A: `cleanup --non-interactive --mode safe` (no --force), run from inside the repo, exits 1 with "destructive operation requires --force with --non-interactive"; the unmerged gone-upstream `feature` branch is preserved. So the non-interactive gate does compel --force.
- TEST B: `cleanup --non-interactive --mode safe --force` exits 0, prints "Deleted 1 branch(es) with gone upstream", and the unmerged `feature` (commit not in main; `git branch --merged main` excludes it) is force-deleted with -D in SAFE mode.
- cleanup.go:117: DeleteGoneBranches runs unconditionally, independent of opts.Mode. branches.go:45-48 upgrades -d to -D when opts.Force. All confirmed.

WHY IT IS REFUTED (strongest first):

1. "Safe mode" does not mean "never force-delete" in this codebase. Mode gates only NON-WORKTREE branch deletion: CleanNonWorktreeBranches early-returns when Mode != Aggressive (branches.go:96). Gone-upstream branch handling (DeleteGoneBranches) runs in BOTH modes by design, and --force is the documented -d/-D selector there. So `safe + force -> -D` on a gone branch is CONSISTENT with the tool's actual definition of "safe," not a violation of it. The reviewer's core premise ("safe should never force-delete") imports a notion of safety the tool never promised.

2. Faithful implementation of the audit's sanctioned Option B. The audit (flow-audit-2026-06-09.md:188) explicitly offered Option A (a separate --force-unmerged flag) OR Option B ("plumb the dispatch-level Force (DispatchResult.Force) into cleanup.Options.Force when the intent is that one --force means both"). main.go:147-151 (non-interactive re-inject) and 198-199 (interactive opts.Force=true) are exactly Option B. The reviewer even concedes the audit named this; that is an admission the behavior is a chosen, documented resolution, not an accidental defect.

3. --force is an explicit user opt-in. A user who types --force getting force-deletion is the plain meaning of the flag.

4. The interactive twin (main.go:198-199) is not "compelled" at all: the gate (registry.go:104) only fires for non-interactive mode, and the TUI confirms each destructive op before deleting. The "compelled to pass --force" complaint applies solely to the non-interactive path, where compulsion is pre-existing gate behavior in registry.go (unchanged by this diff).

HONEST CAVEAT (the legitimate kernel): the fix does remove the ability to run a NON-INTERACTIVE cleanup that cleans gone branches while preserving unmerged ones, because the gate mandates --force and --force now means -D. That is a real capability change versus the pre-fix state (pre-fix, --force was dead and unmerged gone branches were always preserved with -d, with a contradictory "use --force" message). But that is precisely the Option-A-vs-B tradeoff the audit surfaced: someone who wants that capability wants Option A's distinct --force-unmerged flag. It is a design PREFERENCE between two named, acceptable options, not a defect in the implementation of the option that was deliberately chosen.

SEVERITY: N/A as a "bug introduced by the fix" — the behavior matches documented design intent and requires explicit --force opt-in. As a registered design tradeoff it is at most P3.

Files: behavior change lives entirely in /Users/abiswas/code/personal/sentei/fix-detect-default-branch/main.go (lines 146-153 and 198-200). The gate (internal/cli/registry.go:104), the -d/-D upgrade (internal/cleanup/branches.go:45-48), and the unconditional run (internal/cleanup/cleanup.go:117) are all unchanged by this diff and define the "safe mode does not gate gone-branch handling" semantics that make the observed behavior correct-by-design.
- (state-fileutil) RemoveAllRetry spins up to 50 times on permanent errors instead of failing fast
  - The reviewer's factual observations are correct but do not constitute a bug introduced by the fix, and the reviewer explicitly concedes this ("a robustness/clarity issue, not a correctness bug"; "it cannot loop forever and returns the right error").

Confirmed by experiment:
- Permanent error (read-only parent -> EACCES): the loop runs all 50 iterations, then returns the correct last error verbatim ("unlinkat .../sub: permission denied"). No error-type guard, as claimed.
- Success path: returns nil after 1 iteration. Missing path: os.RemoveAll returns nil, so 1 iteration, nil. The bound is enforced; no infinite loop.

Why this is not a bug in the fix:
1. Contract is fully met. The function removes the path and, on a permanent failure, returns the real error. No wrong/inconsistent state, no swallowed error, no regression vs prior behavior (the file is brand new; before this branch clone.go did not even exist).
2. The "wasted work" is negligible and below human perception. Measured: 50 passes on a permanent EACCES took ~2ms for a small tree and ~82ms for a 2000-file tree, because os.RemoveAll fails fast at the first un-removable entry instead of re-walking the entire tree 50 times. So even the worst realistic permanent-error case adds only tens of milliseconds.
3. This only ever runs on an already-failing cleanup/rollback path. In the highest-frequency consumer (clone.go:76) the result is discarded (`_ = ...`), so the extra retries have zero functional consequence. migrate.go still surfaces the correct error to the user.
4. The permanent-error case (EACCES/EPERM/open handle) is uncommon on the target macOS dev workflow, and when it occurs the behavior degrades gracefully with the right error.

The legitimate residue is pure polish: a magic constant 50, no errors.Is(ENOTEMPTY) filter, and a comment that slightly overstates what a sleepless retry buys for a holder needing time to release. None of this rises to a correctness defect or a bug the fix introduced. Verdict: refuted (the area is functionally clean; at most a P3 clarity nit, not a bug).
- (test-infra) playground.Setup() in a TestMain package writes real git objects outside the Spotlight-excluded base, so the flake protection does not reach the very test it should
  - The reviewer's MECHANISM is technically true and I confirmed it empirically: playground.PlaygroundDir = filepath.Join(os.TempDir(), "sentei-playground") is captured at package-init, so even when the tui package's new TestMain sets TMPDIR to a Spotlight-excluded base, playground.Setup() still writes real git objects to the real /var/folders/.../T/sentei-playground (verified by logging it from inside the tui package under TestMain). The tui TestMain's TMPDIR isolation does not reach this test.

However, the FILED BUG (P2: "flake protection does not reach the very test it should") is both pre-existing and functionally unsubstantiated, so it is not a bug introduced by this fix:

1. NOT INTRODUCED BY THE DIFF. The git diff hunks for confirm_test.go contain only the testtmp import and the TestConfirmDeletion_UnlocksLockedWorktrees change (line 89). TestPlayground_DeleteAll_IncludesLockedWorktree (line 160) and playground.Setup() and the PlaygroundDir package var are all pre-existing and untouched. The finding itself concedes this. A review for bugs the fixes introduced requires showing the fix made something worse or left the targeted failure reachable; neither holds.

2. THE ONE IN-SCOPE PLAYGROUND CHANGE IMPROVED FLAKE RESISTANCE. The diff's only change to setup.go is line 111: os.RemoveAll(tmpWork) -> fileutil.RemoveAllRetry(tmpWork) (commit ad5043a "retry seed cleanup to kill the macOS RemoveAll flake"). The authors addressed the playground flake via a bounded retry (the same strategy as testtmp.removeWithRetry), not via TMPDIR isolation.

3. NO PATH CARRIES A SPOTLIGHT REMOVEALL ERROR INTO A TEST FAILURE. The three RemoveAll calls in setup.go: line 38 (pre-wipe, error discarded with `_ =`, followed by MkdirAll which is a no-op on an existing dir); line 45 (cleanupFn, invoked via deferred cleanup(), error discarded); line 111 (the only error-PROPAGATING git-object removal, now retry-protected). So the transient Spotlight ENOTEMPTY race cannot fail this test through any propagating RemoveAll.

4. playground.Setup() CANNOT move under an isolated TMPDIR: it is a production CLI feature (--playground, main.go:295-301) that deliberately uses a fixed, discoverable path and prints "Playground repo: %s". The tui TestMain's TMPDIR isolation correctly targets the OTHER tui tests that use t.TempDir(); it was never required to cover playground.Setup().

Honest caveat (does not change the verdict): lines 38 and 45 use bare os.RemoveAll without retry. In a rare compounded scenario (a prior run's deferred cleanup flaked, leaving worktree dirs, AND the next run's pre-wipe also flaked), a later `git worktree add` could fail. But this is pre-existing (diff touched neither line), error-discarded within any single run, and at most P3 robustness, not the filed P2. It does not make the diff's fix incomplete or wrong.

Verdict: mechanism confirmed true, impact unsubstantiated and pre-existing, the in-scope flake spot is retry-protected. Refuted.
- (test-infra) removeWithRetry does not cover plain t.TempDir() cleanups, so the anti-flake guarantee for TestMain packages rests solely on the .metadata_never_index marker
  - Working-as-designed, not a bug. Two of the finding's sub-claims are factually true but neither produces wrong behavior.

TRUE SUB-CLAIM 1 (retry scope): removeWithRetry's only two call sites are RobustTempDir's t.Cleanup and the post-m.Run() base teardown; it never wraps per-test t.TempDir() dirs, which the framework cleans via testing.go removeAll (testing.go:1392 retries only Windows "Access is denied", gated on GOOS=="windows" in os/removeall_*.go). So on macOS ENOTEMPTY the framework does NOT retry t.TempDir() cleanup. All accurate.

WHY IT IS NOT A BUG: The protection for plain t.TempDir() in the six TestMain packages comes from the MARKER, not the retry. RunWithIsolatedTemp redirects TMPDIR to a base dir carrying .metadata_never_index (a recursive Spotlight subtree exclusion). My live experiment proved t.TempDir() then creates its dirs UNDER that marked base (output: t.TempDir()=.../sentei-tests-NNN/TestX/001 with "FOUND marker at ancestor: .../sentei-tests-NNN"). So those dirs ARE Spotlight-excluded, exactly as the doc states ("every t.TempDir in the package is safe from the macOS indexing race"). The retry not covering t.TempDir() is by design: the marker (the correct documented mechanism) already does. The package doc's "two ways" describes two mechanisms the helper provides, not a claim that both apply to every test; the marker is the load-bearing one for t.TempDir(), the retry is belt-and-suspenders for the helper-owned dirs. The marker is written to base BEFORE m.Run(), so there is no write-ordering race.

TRUE SUB-CLAIM 2 (no-backoff loop): removeWithRetry loops 50x with no sleep/yield, so it absorbs only a sub-microsecond transient. True, but harmless by construction: both call sites swallow the RemoveAll result (no t failure, no os.Exit change). Even a completely no-op retry causes zero wrong behavior. It is at most a minor quality nit, not the claimed P3 defect, and it does not undermine the anti-flake guarantee since the marker is what prevents the hold.

MISATTRIBUTION: The finding cites internal/repo as an example of an unprotected plain-t.TempDir() path. internal/repo is NOT one of the six TestMain packages (it has no testmain_test.go, so it never redirects TMPDIR) and uses RobustTempDir directly in its git-writing e2e_test.go. The example does not support the finding's claim about the six TestMain packages.

NO REPRODUCED WRONG BEHAVIOR: A Spotlight ENOTEMPTY hold cannot be deterministically reproduced, and the finding itself concedes the marker is "plausibly correct." Per the verification mandate (confirm only with a reproduced wrong behavior or a traced consumer that mishandles the case), there is nothing to confirm: the consumer path is correct, the marker covers t.TempDir(), and the harmless retry's only effect is best-effort cleanup of helper-owned dirs.
