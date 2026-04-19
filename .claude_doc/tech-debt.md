# Tech Debt — 未修項目

從 2026-04-19 的 code review 留下的未處理項目。已處理的見 `architecture.md` 的 Concurrency model 段。

## 中（正確性隱患，目前沒炸但脆弱）

### pull stash no-op 判斷依賴英文訊息
- 位置：`internal/pull/ops.go` `cascade`（Phase 2）
- 現況：`stashed := stashErr == nil && !strings.Contains(stashOut, "No local changes")`
- 問題：locale 非英文或 git 版本換訊息時誤判。誤判成 stashed 後，下一步 `stash pop` 會彈出使用者**之前的** stash，造成資料污染。
- 修法：改 `diff-index --quiet HEAD --` 預判 dirty，只在 dirty 時才 stash。

### `pull.conflictFiles` 沒用 -z
- 位置：`internal/pull/ops.go:178`
- 現況：`status --porcelain` + `strings.Split("\n")`
- 問題：檔名含空白/換行會解析錯。
- 修法：改用 `status --porcelain -z`，復用 `internal/branch/ops.go` 的 `parsePorcelainZ`（把它抽到 `util` 或 `gitutil` 共用）。

## 中（安全/race）

### completion cache 路徑可預測
- 位置：`internal/completion/cache.go:47`
- 現況：`os.TempDir()` + sha1(root)，world-writable，其他 user 可搶先建惡意 cache。
- 修法：改 `os.UserCacheDir()`，檔案權限 `0600`。

### completion cache 非原子寫
- 位置：`internal/completion/cache.go:68`
- 現況：`os.WriteFile`，多個 shell tab 同時補完會互相截斷。
- 修法：`os.CreateTemp` + `os.Rename`。

### `repo.FindGitRepos` 排除 worktree / submodule
- 位置：`internal/repo/discovery.go:54`
- 現況：`info.IsDir()` 只接受 `.git` 是目錄的 repo。
- 問題：git worktree 子 repo 和 submodule 的 `.git` 是檔案，會被忽略。
- 決策未定：是 bug 還是 feature？若要支援，用 `gitutil.GitOK(abs, "rev-parse", "--git-dir")` 判定。

## 低（code smell）

### main.go dispatch 沒做完的重構
- 140 行 switch，每個 subcommand 手寫 `validate.BranchName` + for 迴圈。
- 原本 commit message 寫做了 opSpec table，實際只在 completion 路徑做，主 dispatch 還是手寫。
- 修法：每個 subcommand 定義 `{ArgSchema, Runner, Concurrent bool}`，統一跑。

### push 沒併發化
- `main.go:219` 用 for 迴圈序列呼 `push.Push`。
- 和 `fetch`/`status` 風格不一致，沒有好理由序列。push 本身不互動。
- 注意：改併發後輸出要改走 Pattern 3（buffered combined output），見 architecture.md。

### stash list / pop 雙 fork
- `internal/stash/ops.go:58`：先 `stash list` 判斷，再 `stash pop`。
- `git stash pop` 對空 stash 自己會失敗，吞 error 即可省一次 fork。

### 自 parse `.git/config`
- `internal/status/display.go:81` `readOriginURL` 手寫 config parser，還要 fallback 到 `git remote get-url`。
- `git config --get remote.origin.url` 其實很快，且能正確處理 worktree（`.git` 是檔案）。
- 修法：刪 parser，一律用 `git config --get`。

### upstream 偵測邏輯
- `internal/push/ops.go:20` 用 `rev-parse @{u}` 判斷 upstream。
- 改 `for-each-ref --format='%(upstream:short)' refs/heads/<branch>` 語意更明確，branch 名含特殊字元也安全。

### `argOrEmpty` / `containsFlag` 位置
- `main.go:337-353` 留在 main 尾巴，應挪 `internal/util`（或 inline 掉，只有一兩次用）。

### 表寬硬編
- `fetch/ops.go` 和 `status/display.go` 各自硬編欄寬（`%-20s %-30s` vs 動態 `labelW/branchW`）。
- 可抽 `ui.Table` helper。目前兩處不同風格看起來不一致。

### `ListAllNames` slice-of-slices
- `internal/branch/ops.go:134` 已改用 `ParallelMap`，但仍是 `[][]string` 收集後 flatten + dedupe。
- 若想純淨換 `chan []string` + single collector。極低優先。

## IDE 靜態提示

- `strings.SplitSeq`：go 1.24 加的 iterator 版本，更有效率。多處可換（`branch/ops.go`、`stash/ops.go`、`pull/ops.go` 等用 `strings.Split(...)` 跑 for-range 的地方）。
- cSpell：`serialises`、`ansi` 等拼字警告，噪音，加到字典或忽略。
