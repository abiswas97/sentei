# Changelog

## [1.4.0](https://github.com/abiswas97/sentei/compare/v1.3.0...v1.4.0) (2026-04-01)


### Features

* activate playground min-progress-duration (1.5s hold) ([e3ee831](https://github.com/abiswas97/sentei/commit/e3ee83104499d798affa4ba8a6901a376ad077ca))
* add --dry-run support to remove command ([0a140c8](https://github.com/abiswas97/sentei/commit/0a140c860fa8bebd20153a04bdffd8aa3c660aaa))
* add cleanup confirmation view and TUI entry for decision commands ([5a52383](https://github.com/abiswas97/sentei/commit/5a5238334d91be413c485f421305982be8671986))
* add clone command with --non-interactive support ([7c54ab0](https://github.com/abiswas97/sentei/commit/7c54ab05849be3cbfc67e2a0eab5d700437967fa))
* add command registry for CLI/TUI handoff ([638e6d2](https://github.com/abiswas97/sentei/commit/638e6d2854e04641438acb3481c6a5787f528db5))
* add create worktree command with --non-interactive support ([1fb2556](https://github.com/abiswas97/sentei/commit/1fb25563ca6ed457b4005ffec79b80295fbd7eb0))
* add holdOrAdvance helper and progressHoldExpiredMsg for min-duration hold ([a578e83](https://github.com/abiswas97/sentei/commit/a578e83083fd323391bddff0b6bb3a5d04da1a87))
* add migrate command with --non-interactive support ([968134c](https://github.com/abiswas97/sentei/commit/968134c1bb922e8aed22e9d6789359525bc78476))
* add ModelOption + WithMinProgressDuration to tui.Model ([8fa5bf2](https://github.com/abiswas97/sentei/commit/8fa5bf292ea9bdc86e2cf82f921756351a0b6b61))
* add progress tracker package, fix integration progress to use upfront total ([bfd8130](https://github.com/abiswas97/sentei/commit/bfd8130ee75b4373597edce30e6bd04c7e19f50e))
* add remove command with filter flags and --non-interactive ([efa4108](https://github.com/abiswas97/sentei/commit/efa41087a697d34024ca4d06cdbbe453d1356af5))
* add remove filter pre-selection and TUI tests ([4e4e2d0](https://github.com/abiswas97/sentei/commit/4e4e2d002e23a4fa5a045ac85908882766b1e9fb))
* add reusable confirmation view component ([3008561](https://github.com/abiswas97/sentei/commit/30085616fa10889d65b976a790373217c39942ef))
* add teatest infrastructure for TUI E2E testing ([ad31b0e](https://github.com/abiswas97/sentei/commit/ad31b0e8d13346b034b426ca6c1af61369ccf252))
* add TUI confirmation views for create, clone, and migrate ([a37091f](https://github.com/abiswas97/sentei/commit/a37091ff8cb542c8f7c9139576bd1250982a7965))
* add UnlockWorktree for safe locked worktree deletion ([4c1b048](https://github.com/abiswas97/sentei/commit/4c1b048fa56eb08e0b88912f6c596a51bcb3c577))
* CLI/TUI handoff with command registry and --non-interactive support ([06be0b8](https://github.com/abiswas97/sentei/commit/06be0b8c61fa132fa1cc8ea0f498305cbde45801))
* record progress view start time at all entry points ([665449f](https://github.com/abiswas97/sentei/commit/665449fe982d417b17f39b7e34883827f6aba7bd))
* show percentage instead of step counts in all progress views ([91445ee](https://github.com/abiswas97/sentei/commit/91445ee9e17b9489a606955623671c6b21e024e9))
* UI alignment - locked deletion, progress hold, eager reload ([779aa0d](https://github.com/abiswas97/sentei/commit/779aa0d045f0e0aa90745489a7e653b3e8873eca))
* use holdOrAdvance in all progress completion handlers ([686d99f](https://github.com/abiswas97/sentei/commit/686d99ff58f5f560e4b165b7236a10011d8ce8a1))
* wire cleanup command into registry with --non-interactive support ([d178e8d](https://github.com/abiswas97/sentei/commit/d178e8dc357426f559d01998b5237a43c8855a68))


### Bug Fixes

* add Ctrl+C/q handler to all progress views, refresh worktree list ([72f36ce](https://github.com/abiswas97/sentei/commit/72f36cefa574f7a17c1fc70e383e74f506550119))
* address all remaining review suggestions ([3012140](https://github.com/abiswas97/sentei/commit/30121407259355241559bb7844fe0e2c26d98d99))
* address all review findings from final review round ([a9048c7](https://github.com/abiswas97/sentei/commit/a9048c78d52acfe5e802a88e09949a782fccb330))
* address code review findings from codex/gemini ([e4e40ec](https://github.com/abiswas97/sentei/commit/e4e40ec79a4d3a49dfb924901d8aea032136c780))
* address codex final review — remove architectural inversion and dead code ([4ee7922](https://github.com/abiswas97/sentei/commit/4ee79228c38394e25f218d9de1bb73f4473d8b30))
* address verification findings — routing, dead code, determinism ([f84c7a3](https://github.com/abiswas97/sentei/commit/f84c7a3f41833c38cc112fa52f4e975e73cda870))
* cleanup now prunes stale worktrees, TUI filters prunable entries ([d578757](https://github.com/abiswas97/sentei/commit/d578757158c7f16535f40cde045672a85c19fe74))
* eliminate all time.Sleep from tests, enforce condition-based waiting ([717efe0](https://github.com/abiswas97/sentei/commit/717efe04b78c7e29e5bfe38e95a8cb8ffda9fb5d))
* emit StatusSkipped for steps not run, fixing progress bar total ([ee8295d](https://github.com/abiswas97/sentei/commit/ee8295d815dc603cdd7e5c89cc82b95c68e246f4))
* include locked worktrees in CLI deletion with unlock-first strategy ([175e4a5](https://github.com/abiswas97/sentei/commit/175e4a5993d29f214c69de2da0796978c44c264d))
* integration progress bar counts unique steps, not raw events ([0f37618](https://github.com/abiswas97/sentei/commit/0f37618ceed7b87c805e361d1a5061aa050921a7))
* isolate playground tests with per-test temp dirs ([b301fe5](https://github.com/abiswas97/sentei/commit/b301fe5b21b772004630dc66a15f92ac8fdc7d31))
* refresh worktree list when returning to menu from deletion summary ([f883965](https://github.com/abiswas97/sentei/commit/f883965cb45a5baef73705dfd9b4d15f07b32261))
* replace lazy stateStale reload with eager worktree context refresh ([dd8293c](https://github.com/abiswas97/sentei/commit/dd8293ca307fc96a6e9d2086a2c65834159a2efe))
* resolveBareRoot returns correct path for plain bare repos ([34a21c0](https://github.com/abiswas97/sentei/commit/34a21c0e6865c30d6e33f41401db64776848dbb0))
* restore --force flag and repo path for cleanup, fix magic constant ([5c55522](https://github.com/abiswas97/sentei/commit/5c5552221690197a53553ff9f586ac6845c23502))
* set git user identity in confirm test for CI ([5abf438](https://github.com/abiswas97/sentei/commit/5abf438bdab08c59ae33c92bafe338ed1ddfa8c2))
* set git user identity in unlock worktree tests for CI ([6e187ca](https://github.com/abiswas97/sentei/commit/6e187ca089d4848df326ec84d541e6a3e85155ca))
* skip remote ref pruning when origin remote doesn't exist ([d6cc1be](https://github.com/abiswas97/sentei/commit/d6cc1be3daac7d1956ba1ff6dac9788273df6453))
* unlock locked worktrees before deletion in TUI ([6b489f1](https://github.com/abiswas97/sentei/commit/6b489f11a90c0883531b9fe8d453b1e40fd43847))
* use --initial-branch=main in all test bare repo setups ([9e0e2e0](https://github.com/abiswas97/sentei/commit/9e0e2e0a28e164d9094b1dd54281b035cc9ec074))

## [1.3.0](https://github.com/abiswas97/sentei/compare/v1.2.0...v1.3.0) (2026-03-30)


### Features

* add ccc copy optimization for cross-worktree index reuse ([6777dbc](https://github.com/abiswas97/sentei/commit/6777dbcc1f2432a9599366ecabda96f047b7cbae))
* add cleanup results screen with detailed metrics ([9525ee3](https://github.com/abiswas97/sentei/commit/9525ee3cc735a20fa89ff5145f5c9c164caf8dc5))
* add IndexCopyDir field for config-driven index copy optimization ([7f00df7](https://github.com/abiswas97/sentei/commit/7f00df7043278d6ee7c6a99bfd1ee47a3300889b))
* add integration manager for enable/disable across worktrees ([a6a27af](https://github.com/abiswas97/sentei/commit/a6a27af856318524d2b3e4f655fa8490e6594bc8))
* add integration selection during migration onboarding ([df0c4cc](https://github.com/abiswas97/sentei/commit/df0c4cce2b9fbfa5a4995b69412417786a53ac23))
* add Manage integrations menu item with state loading ([1929389](https://github.com/abiswas97/sentei/commit/192938925bb3f470076a701f6aebaff2c8388ade))
* add navigation keys and staged styles for integration views ([ba9ef16](https://github.com/abiswas97/sentei/commit/ba9ef1666d53054936f5079fdcd1f6a61b92de53))
* add on-disk integration artifact detection ([80acf6b](https://github.com/abiswas97/sentei/commit/80acf6bf69bc0089eaf30a289e3206241f8241fc))
* add state persistence for bare repo integration config ([605391c](https://github.com/abiswas97/sentei/commit/605391c250e5638b8de2c9cf77d8a9ed868ffa24))
* **cli:** add 'sentei ecosystems' and 'sentei integrations' subcommands ([a34fc28](https://github.com/abiswas97/sentei/commit/a34fc285f36cb306a4c7afeecc083b571e6bfd9d))
* **config:** add Config types with YAML unmarshaling and tests ([0396b91](https://github.com/abiswas97/sentei/commit/0396b91d39d95b1da18cfd5c0c7c3a4a96d9e211))
* **config:** add embedded default ecosystem definitions (16 ecosystems) ([7261471](https://github.com/abiswas97/sentei/commit/7261471ed5bd8ec2ba91935406c715066e2e484d))
* **config:** implement three-layer config loading with merge and validation ([39ccb72](https://github.com/abiswas97/sentei/commit/39ccb72540a2fc45c6d0942e886273732567bea1))
* **creator:** add ecosystem dependency installation with parallel workspace support ([43d572b](https://github.com/abiswas97/sentei/commit/43d572b8c111d87628ee5c71495f171c1e6da85f))
* **creator:** add integration dependency resolution and setup ([a8f4c08](https://github.com/abiswas97/sentei/commit/a8f4c08b42ced9f24a9cb297ea40c8aacec4dbee))
* **creator:** add integration teardown (scan artifacts, cleanup, fallback deletion) ([df26e13](https://github.com/abiswas97/sentei/commit/df26e139a955b09aeb7842eeaf5ea231d5fd7fd9))
* **creator:** add pipeline types and setup phase (create worktree, merge, copy env) ([9eec78b](https://github.com/abiswas97/sentei/commit/9eec78b1f3379e4166cb7d2b2797b7120fffba48))
* **creator:** wire Run() orchestrator with setup, deps, and integrations phases ([b4b8f88](https://github.com/abiswas97/sentei/commit/b4b8f887ff5fd2116d57525babb40578b13dd0aa))
* **ecosystem:** add registry with priority-ordered detection and glob support ([f853c86](https://github.com/abiswas97/sentei/commit/f853c86f754c34257a50df6556867a86fdb52dd2))
* **ecosystem:** add workspace detection for pnpm, npm, go, and cargo ([6105365](https://github.com/abiswas97/sentei/commit/61053657d0b4a16e5ae225e890cbf786230f7134))
* expand sentei to full worktree lifecycle manager ([e5c0882](https://github.com/abiswas97/sentei/commit/e5c0882d5a770f52704bd057e1beab04ff459a69))
* implement integration apply progress view ([f1fd1b7](https://github.com/abiswas97/sentei/commit/f1fd1b7dd97fbbb4f616e1fe4a443e30a09a401d))
* implement integration list view with toggle and info carousel ([324f484](https://github.com/abiswas97/sentei/commit/324f484931f1487dfc9ccc3bcfaac382994a7bc4))
* **integration:** add registry with code-review-graph and cocoindex-code definitions ([23a4fe0](https://github.com/abiswas97/sentei/commit/23a4fe07284018ac4f820d4f0a5a622b6ca81302))
* redesign integration info dialog with visual hierarchy ([9b616e7](https://github.com/abiswas97/sentei/commit/9b616e72b2d6d2c319c35f7661d2e5b6ccb5b2b9))
* replace integration toggles with repo-level state in create flow ([5efac03](https://github.com/abiswas97/sentei/commit/5efac031daf81bcd0efbbbae58528d3608ce5689))
* repo-level integration management with migration onboarding ([fdb0977](https://github.com/abiswas97/sentei/commit/fdb0977845bdffe99de721acf3643c02bd0af1d2))
* **repo:** add clone repo pipeline with URL-to-name derivation and default branch detection ([94a882b](https://github.com/abiswas97/sentei/commit/94a882bdad0f3cdd35977ef96dcd4f8ead3cf94e))
* **repo:** add create repo pipeline with setup and GitHub phases ([291a911](https://github.com/abiswas97/sentei/commit/291a9113afbfb083c97ba1a1eecd25609de86b78))
* **repo:** add migrate repo pipeline with backup, bare conversion, and file copy ([f254563](https://github.com/abiswas97/sentei/commit/f254563e2b0f9e4d01be7d75b9eb688ba5e48a94))
* **repo:** add shared types and context detection (bare/non-bare/no-repo) ([7af785f](https://github.com/abiswas97/sentei/commit/7af785fc72d78f26e050386dd20982f5d91a690b))
* scaffold integration view states and model ([f078af8](https://github.com/abiswas97/sentei/commit/f078af802d733e378dc1b3e05373c81790d765e7))
* **tui:** adapt menu based on repo context (bare/non-bare/no-repo) ([35bb61e](https://github.com/abiswas97/sentei/commit/35bb61ebf7bd37946cb17535a5873c66bd693aad))
* **tui:** add clone URL input view with real-time name derivation ([08b35d6](https://github.com/abiswas97/sentei/commit/08b35d643b7a5bb45600052c95261c164e6924a2))
* **tui:** add create repo name input and options views with GitHub disclosure ([4182e87](https://github.com/abiswas97/sentei/commit/4182e871960c0c0beeef0997ffe214ef27272e21))
* **tui:** add migrate confirmation and summary views with backup cleanup ([3cd12d7](https://github.com/abiswas97/sentei/commit/3cd12d7412471accb3caeb3c25b6c2b2a6b8626a))
* **tui:** add phase indicator and separator styles for new visual language ([90bb534](https://github.com/abiswas97/sentei/commit/90bb534dac305d4c779fd90ee90ae7928d0a5e88))
* **tui:** add shared progress and summary views with sentei re-launch ([247d6e1](https://github.com/abiswas97/sentei/commit/247d6e13575e207920d647ddc3e377d3719e2b47))
* **tui:** enhance confirm view with integration teardown info and execution ([a44faf8](https://github.com/abiswas97/sentei/commit/a44faf8bc19da547228a0ae2aaa6f034e31d2a37))
* **tui:** enhance removal progress with phased teardown/remove/prune reporting ([a4ab530](https://github.com/abiswas97/sentei/commit/a4ab530ff7f0b47f6bd2f9361646c1993bd172be))
* **tui:** implement create branch view with input validation ([5c3e569](https://github.com/abiswas97/sentei/commit/5c3e569e645162f7edcd67df1557c0ef12ad77e9))
* **tui:** implement create options view with ecosystem and integration toggles ([d196056](https://github.com/abiswas97/sentei/commit/d196056c934656660d6eb0fe6b170ea42a9b6ca7))
* **tui:** implement create progress view with phased parallel reporting ([a7c804b](https://github.com/abiswas97/sentei/commit/a7c804b24c205f37952d4455056d52a8d3bfb761))
* **tui:** implement create summary view with success/failure display ([d20647d](https://github.com/abiswas97/sentei/commit/d20647daec77eaed67191d08021570b0d8d0dc06))
* **tui:** implement menu view as new entry point with lazy worktree loading ([8676518](https://github.com/abiswas97/sentei/commit/867651875983b2877d6dcc9ec24f3b3fa088e59c))
* **tui:** split migrate summary into backup decision + what-next screens ([cb5e807](https://github.com/abiswas97/sentei/commit/cb5e80765046c4893b2d84e23cc0c57c862175bc))
* update main.go to start at menu with lazy worktree loading and config ([174bbca](https://github.com/abiswas97/sentei/commit/174bbca4a41ae2b4043a60350eb1f301f801943a))
* wire context detection into main.go and pass to TUI ([6beed65](https://github.com/abiswas97/sentei/commit/6beed656f4ae7a8dc176a940ae9d5f291d3ac986))


### Bug Fixes

* address code review findings (validation, imports, detect output) ([b0c74d9](https://github.com/abiswas97/sentei/commit/b0c74d9e192695ece03482415d6e091d2b7a7754))
* address critical review findings (shell runner, binary detection, validation) ([52fdfb6](https://github.com/abiswas97/sentei/commit/52fdfb6bc20b19d701f556e025c8d0543f9dc41c))
* address final review findings (shell injection, dedup, shared utilities) ([2a75385](https://github.com/abiswas97/sentei/commit/2a75385413b4559d424cbd258f165cf7af80de73))
* address remaining review findings (result propagation, async menu, parallel teardown, registry pattern) ([81611f8](https://github.com/abiswas97/sentei/commit/81611f81bc205d5c010c99e53f9f474bcae0b691))
* address review findings in implementation plan ([45602e5](https://github.com/abiswas97/sentei/commit/45602e5111540c1ea5dbeb99f22d366f7034e7c7))
* allow esc to exit from main menu ([582c0e9](https://github.com/abiswas97/sentei/commit/582c0e90344167c1465bb2dbf31fe8e883bfedd5))
* auto-update location as repo name is typed, use repoPath not CWD ([0261fe8](https://github.com/abiswas97/sentei/commit/0261fe8ee83b15e58936f941e7f5f2d9b315656d))
* clean root directory after migration, add E2E assertion for clean root ([7049f73](https://github.com/abiswas97/sentei/commit/7049f73e10555625028dadd34afd3047641d8912))
* enable worktree toggle, focus description input on cursor navigation ([f31cbaf](https://github.com/abiswas97/sentei/commit/f31cbaf96a5782290c216f76c2384c027bdc095d))
* make info dialog responsive to terminal width with text wrapping ([821354c](https://github.com/abiswas97/sentei/commit/821354c191a161766b1a4e234815a77171cd04db))
* make integration descriptions more descriptive for info dialog ([aaf6051](https://github.com/abiswas97/sentei/commit/aaf6051c398f0fb52e5847bd292abb6abd7d92f4))
* options toggle, description input, gh create args, error wrapping ([f1e14d5](https://github.com/abiswas97/sentei/commit/f1e14d5f0004cc9761776ba44dbfbeff0855c607))
* resolve absolute paths and pre-fill CWD in create repo location ([ccb8426](https://github.com/abiswas97/sentei/commit/ccb8426f138ef34ffc58e43fde68113610614155))
* resolve bare repo root from inside worktrees using --git-common-dir ([769842b](https://github.com/abiswas97/sentei/commit/769842bbc1bd5bfd46c0e9db65386f6637980dc7))
* resolve CI failures — lint QF1012 and race in parallel deps ([5967973](https://github.com/abiswas97/sentei/commit/59679734d8d443ab54dbcfd43b5b41db7f0bd5f6))
* standardize key bindings to hjkl throughout, fix title consistency ([0901eab](https://github.com/abiswas97/sentei/commit/0901eab81fe32e53850bafd24aef73381880a033))
* standardize remove worktrees flow UI to match app-wide style ([242dfe5](https://github.com/abiswas97/sentei/commit/242dfe5151b35e14b06a49f44c04d0f5b2efda81))
* update post-migration text to remove specific integration names ([4e5460a](https://github.com/abiswas97/sentei/commit/4e5460af3806d390f2bd7c8fde310699fe38193b))

## [1.2.0](https://github.com/abiswas97/sentei/compare/v1.1.0...v1.2.0) (2026-03-21)


### Features

* add esc to quit keys ([af6c5f6](https://github.com/abiswas97/sentei/commit/af6c5f6276502514728d9b54d5d06e023fb3b9a7))
* add version and readme ([51ea174](https://github.com/abiswas97/sentei/commit/51ea1740d984c99de78f9a5ba701d614fba228ee))
* auto prune orphans ([9350f13](https://github.com/abiswas97/sentei/commit/9350f134a2e146ba8a3d31494bac8b2a6875b00a))
* branch protection ([00df628](https://github.com/abiswas97/sentei/commit/00df628d8568d4860e74495815b64a7d2e7e6e44))
* cicd pipeline and releases ([e4a7a64](https://github.com/abiswas97/sentei/commit/e4a7a646457617ad9dcd41c013734a407c88255c))
* cicd pipeline and releases ([8d2d8b8](https://github.com/abiswas97/sentei/commit/8d2d8b8b3bef2ad716cd696d992ed87daadc436f))
* clean up ui to be more tabular ([9a3c4a7](https://github.com/abiswas97/sentei/commit/9a3c4a7e2d7b3498de55fe90f70fcf223ecc3aea))
* **cleanup:** add CLI subcommand and TUI integration ([4c1fca0](https://github.com/abiswas97/sentei/commit/4c1fca04b48b889196b9773bdaaf18d579fa767d))
* **cleanup:** add types, fixtures, resolveConfigPath test, and stub orchestrator ([a9debb1](https://github.com/abiswas97/sentei/commit/a9debb161b7c4fce845092477481c60851f4c016))
* **cleanup:** implement branch cleanup (gone-upstream and non-worktree) ([4fc4272](https://github.com/abiswas97/sentei/commit/4fc427209132d2efa4c028a7352d7d5a66b98bc6))
* **cleanup:** implement config dedup with atomic writes ([8db2f74](https://github.com/abiswas97/sentei/commit/8db2f7408a818dee01e8401d3f5672147507655e))
* **cleanup:** implement orphaned config section purge ([a1f8964](https://github.com/abiswas97/sentei/commit/a1f896495485465b5c7304e400217f5ff19febd1))
* **cleanup:** implement remote ref pruning ([26ba91b](https://github.com/abiswas97/sentei/commit/26ba91bcb28f390016d522ead47a0f06162e40ce))
* **cleanup:** wire orchestrator with all 5 pipeline steps ([3e027e9](https://github.com/abiswas97/sentei/commit/3e027e9dc230ce372a4d346ec3ed88db41fbf15d))
* dry run mode ([881bc99](https://github.com/abiswas97/sentei/commit/881bc999ba22f84203f00843284a5be2099b43b3))
* init ([408af98](https://github.com/abiswas97/sentei/commit/408af987b682ae1cd407a1e931d3ad42579bd388))
* interactive test playground ([0362dc7](https://github.com/abiswas97/sentei/commit/0362dc74431155213ae7c377c444bb8b59b4964e))
* PRD ([c4779b1](https://github.com/abiswas97/sentei/commit/c4779b162c8101c5e8c8c022a21f198ce44a635e))
* progress indicator ([a1b78fe](https://github.com/abiswas97/sentei/commit/a1b78fec8df784da824ab63658b7e79a9627dfd7))
* sort and filter ([4008dfa](https://github.com/abiswas97/sentei/commit/4008dfa02f3e2fcb0d4f36b4d783f1327ef71dea))
* status legend ([fe36b75](https://github.com/abiswas97/sentei/commit/fe36b7535b11925bf68595fcb0cee66d90237b08))
* ui setup ([ac02c5c](https://github.com/abiswas97/sentei/commit/ac02c5c55886700a0521e22d862112d21ced94a9))
* worktree enricher ([ba61c50](https://github.com/abiswas97/sentei/commit/ba61c500d1ba52deebf71186b07c6474191612dc))
* worktree parser ([dcf278b](https://github.com/abiswas97/sentei/commit/dcf278b63fe682fa21f24cf1dd38382a2496f740))


### Bug Fixes

* **cleanup:** address code review findings ([3e3c7bd](https://github.com/abiswas97/sentei/commit/3e3c7bd283e34467124610461fd1d7aa09dd104a))
* code review ([2833e16](https://github.com/abiswas97/sentei/commit/2833e168690d23ef854100e823c4e3aca481dcfe))
* merge release workflows so goreleaser triggers on release ([15834a0](https://github.com/abiswas97/sentei/commit/15834a07f22a56bd2451c7f8d1c3f83595db30de))
* merge release workflows so goreleaser triggers on release ([ef9c46f](https://github.com/abiswas97/sentei/commit/ef9c46f4fe9e9797e5e72bcd439fd518212957f7))
* potential guard ([5bfc617](https://github.com/abiswas97/sentei/commit/5bfc6179e3529d872bb9ec49d127f61a92e67a5e))
* remaining staticcheck QF1012 lint errors in summary.go ([640d4b1](https://github.com/abiswas97/sentei/commit/640d4b1d74c20b169a22e33f542409e58c3a3ff6))
* remove component prefix from release-please tags ([22257a7](https://github.com/abiswas97/sentei/commit/22257a74ea9f09ae95add086205e953afa9baba4))
* remove component prefix from release-please tags ([9e1e67a](https://github.com/abiswas97/sentei/commit/9e1e67a8e54559cfb5f75082f752317e7fa977a9))
* replace git worktree remove with os.RemoveAll for reliable cleanup ([be03198](https://github.com/abiswas97/sentei/commit/be031988be2bccd79ca6ee30b0787b6c76c5d612))
* resolve lint errors (errcheck, staticcheck QF1012) ([ae27fd7](https://github.com/abiswas97/sentei/commit/ae27fd7ae1470b8f46063bd0d8e4cf23c0a2ee60))
* ubuntu test failure ([3a0ef94](https://github.com/abiswas97/sentei/commit/3a0ef94e3fda359777454a184e72bf2fe753b3ac))
* use fmt.Fprintf instead of WriteString(fmt.Sprintf) to satisfy staticcheck QF1012 ([daa1e2d](https://github.com/abiswas97/sentei/commit/daa1e2dade5bb3a6d32e9779ba5bd3202f53f9ee))
* use homebrew formula instead of cask and reset version ([9f8fcae](https://github.com/abiswas97/sentei/commit/9f8fcaeb6f45239bb313f01943539d3fd27b4ad8))
* use homebrew formula instead of cask and reset version ([023a717](https://github.com/abiswas97/sentei/commit/023a717de479a879256728b553af29800d132e09))
* use PAT for release-please to trigger CI on PR branches ([50225fd](https://github.com/abiswas97/sentei/commit/50225fdc845e1a7d130edfde48b3102d5d73d72f))

## [1.1.0](https://github.com/abiswas97/sentei/compare/v1.0.1...v1.1.0) (2026-03-21)


### Features

* **cleanup:** add CLI subcommand and TUI integration ([4c1fca0](https://github.com/abiswas97/sentei/commit/4c1fca04b48b889196b9773bdaaf18d579fa767d))
* **cleanup:** add types, fixtures, resolveConfigPath test, and stub orchestrator ([a9debb1](https://github.com/abiswas97/sentei/commit/a9debb161b7c4fce845092477481c60851f4c016))
* **cleanup:** implement branch cleanup (gone-upstream and non-worktree) ([4fc4272](https://github.com/abiswas97/sentei/commit/4fc427209132d2efa4c028a7352d7d5a66b98bc6))
* **cleanup:** implement config dedup with atomic writes ([8db2f74](https://github.com/abiswas97/sentei/commit/8db2f7408a818dee01e8401d3f5672147507655e))
* **cleanup:** implement orphaned config section purge ([a1f8964](https://github.com/abiswas97/sentei/commit/a1f896495485465b5c7304e400217f5ff19febd1))
* **cleanup:** implement remote ref pruning ([26ba91b](https://github.com/abiswas97/sentei/commit/26ba91bcb28f390016d522ead47a0f06162e40ce))
* **cleanup:** wire orchestrator with all 5 pipeline steps ([3e027e9](https://github.com/abiswas97/sentei/commit/3e027e9dc230ce372a4d346ec3ed88db41fbf15d))


### Bug Fixes

* **cleanup:** address code review findings ([3e3c7bd](https://github.com/abiswas97/sentei/commit/3e3c7bd283e34467124610461fd1d7aa09dd104a))
* resolve lint errors (errcheck, staticcheck QF1012) ([ae27fd7](https://github.com/abiswas97/sentei/commit/ae27fd7ae1470b8f46063bd0d8e4cf23c0a2ee60))

## [1.0.1](https://github.com/abiswas97/sentei/compare/v1.0.0...v1.0.1) (2026-03-05)


### Bug Fixes

* replace git worktree remove with os.RemoveAll for reliable cleanup ([be03198](https://github.com/abiswas97/sentei/commit/be031988be2bccd79ca6ee30b0787b6c76c5d612))
* use fmt.Fprintf instead of WriteString(fmt.Sprintf) to satisfy staticcheck QF1012 ([daa1e2d](https://github.com/abiswas97/sentei/commit/daa1e2dade5bb3a6d32e9779ba5bd3202f53f9ee))

## 1.0.0 (2026-02-07)


### Features

* add esc to quit keys ([af6c5f6](https://github.com/abiswas97/sentei/commit/af6c5f6276502514728d9b54d5d06e023fb3b9a7))
* add version and readme ([51ea174](https://github.com/abiswas97/sentei/commit/51ea1740d984c99de78f9a5ba701d614fba228ee))
* auto prune orphans ([9350f13](https://github.com/abiswas97/sentei/commit/9350f134a2e146ba8a3d31494bac8b2a6875b00a))
* branch protection ([00df628](https://github.com/abiswas97/sentei/commit/00df628d8568d4860e74495815b64a7d2e7e6e44))
* cicd pipeline and releases ([e4a7a64](https://github.com/abiswas97/sentei/commit/e4a7a646457617ad9dcd41c013734a407c88255c))
* cicd pipeline and releases ([8d2d8b8](https://github.com/abiswas97/sentei/commit/8d2d8b8b3bef2ad716cd696d992ed87daadc436f))
* clean up ui to be more tabular ([9a3c4a7](https://github.com/abiswas97/sentei/commit/9a3c4a7e2d7b3498de55fe90f70fcf223ecc3aea))
* dry run mode ([881bc99](https://github.com/abiswas97/sentei/commit/881bc999ba22f84203f00843284a5be2099b43b3))
* init ([408af98](https://github.com/abiswas97/sentei/commit/408af987b682ae1cd407a1e931d3ad42579bd388))
* interactive test playground ([0362dc7](https://github.com/abiswas97/sentei/commit/0362dc74431155213ae7c377c444bb8b59b4964e))
* PRD ([c4779b1](https://github.com/abiswas97/sentei/commit/c4779b162c8101c5e8c8c022a21f198ce44a635e))
* progress indicator ([a1b78fe](https://github.com/abiswas97/sentei/commit/a1b78fec8df784da824ab63658b7e79a9627dfd7))
* sort and filter ([4008dfa](https://github.com/abiswas97/sentei/commit/4008dfa02f3e2fcb0d4f36b4d783f1327ef71dea))
* status legend ([fe36b75](https://github.com/abiswas97/sentei/commit/fe36b7535b11925bf68595fcb0cee66d90237b08))
* ui setup ([ac02c5c](https://github.com/abiswas97/sentei/commit/ac02c5c55886700a0521e22d862112d21ced94a9))
* worktree enricher ([ba61c50](https://github.com/abiswas97/sentei/commit/ba61c500d1ba52deebf71186b07c6474191612dc))
* worktree parser ([dcf278b](https://github.com/abiswas97/sentei/commit/dcf278b63fe682fa21f24cf1dd38382a2496f740))


### Bug Fixes

* code review ([2833e16](https://github.com/abiswas97/sentei/commit/2833e168690d23ef854100e823c4e3aca481dcfe))
* merge release workflows so goreleaser triggers on release ([15834a0](https://github.com/abiswas97/sentei/commit/15834a07f22a56bd2451c7f8d1c3f83595db30de))
* merge release workflows so goreleaser triggers on release ([ef9c46f](https://github.com/abiswas97/sentei/commit/ef9c46f4fe9e9797e5e72bcd439fd518212957f7))
* potential guard ([5bfc617](https://github.com/abiswas97/sentei/commit/5bfc6179e3529d872bb9ec49d127f61a92e67a5e))
* remove component prefix from release-please tags ([22257a7](https://github.com/abiswas97/sentei/commit/22257a74ea9f09ae95add086205e953afa9baba4))
* remove component prefix from release-please tags ([9e1e67a](https://github.com/abiswas97/sentei/commit/9e1e67a8e54559cfb5f75082f752317e7fa977a9))
* ubuntu test failure ([3a0ef94](https://github.com/abiswas97/sentei/commit/3a0ef94e3fda359777454a184e72bf2fe753b3ac))
* use homebrew formula instead of cask and reset version ([9f8fcae](https://github.com/abiswas97/sentei/commit/9f8fcaeb6f45239bb313f01943539d3fd27b4ad8))
* use homebrew formula instead of cask and reset version ([023a717](https://github.com/abiswas97/sentei/commit/023a717de479a879256728b553af29800d132e09))

## Changelog
