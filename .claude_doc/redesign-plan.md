# gitmulti Redesign Plan

## 命名改變：flag → subcommand

| 現在 | 新命名 |
|------|--------|
| `-p [branch]` | `pull [branch]` |
| `-s <branch>` | `switch <branch>` |
| `-sf <branch>` | `switch -f <branch>` |
| `-nb <branch>` | `switch -c <branch>` |
| `-st` | `status` |
| `-b` | `branch` |
| `-al [kw]` | `branch -a [kw]` |
| `-F <kw>` | `branch --find <kw>` |
| `-dc` | `discard` |
| `-d <path>` | `-C <path>` |

---

## 指令行為規格

### `pull [branch]`
三段 cascade，每步有輸出：

1. 嘗試 `git pull --ff-only`
   - 成功 → 結束
2. 失敗 → 印出 `→ ff-only failed, stashing changes...`，執行 stash → pull → stash pop
   - stash pop 成功 → 該 repo 結束
3. 還是失敗 → 所有衝突 repo 一起跳出選單（一次套用）：
   ```
   repo-a: conflict after pull
   repo-b: conflict after pull

   All conflicted repos:
     1) reset --hard (丟棄本地變更)
     2) merge
     3) rebase
     4) skip all
   ```
   選 2/3 → 印出衝突檔案 + 提示手動指令後結束：
   ```
   repo-a: rebase conflict
     M src/App.tsx
     → resolve manually, then run: git rebase --continue
   ```

### `pull --rebase [branch]`
直接執行 `git pull --rebase`，不走 cascade。

### `push [branch]`
- reject → 報錯停下，讓用戶自己處理
- 新 branch（無 upstream）→ 自動執行 `git push -u origin <branch>`，不詢問

---

### `fetch`
執行 `git fetch` 後自動印出 ahead/behind 狀態：
```
frontend   feature/login   ↑3 ↓2   dirty
backend    main            ↑0 ↓1   clean
shared     main            ↑0 ↓0   clean
```

---

### `switch <branch>`
有 uncommitted changes 時提示確認。

### `switch -f <branch>`
強制切換，不提示。

### `switch -c <branch>`
對所有 repo 建立新 branch。

---

### `status`
顯示檔案層級變更（只顯示有變更的 repo）：
```
frontend:
  M src/App.tsx
  ?? newfile.txt

backend:
  M routes/user.go
```

### `branch`
顯示 repo 層級 ahead/behind 狀態：
```
frontend   feature/login   ↑3 ↓0   dirty
backend    main            ↑0 ↓2   clean
shared     main            ↑0 ↓0   clean
```

### `branch -a [keyword]`
列出所有 unique branch，可選 keyword 過濾。

### `branch --find <keyword>`
跨 repo 找符合 keyword 的 branch。

---

### `branch -d <name>`
1. 掃描哪些 repo 有該 branch，沒有的給提示跳過
2. 列出操作清單讓用戶確認：
   ```
   Deleting branch: feature/login

     frontend   has branch
     backend    has branch
     shared     no branch → skip

   Delete in frontend, backend?
     1) yes to all
     2) confirm each
   ```
3. 支援 `--remote` 同時刪除 remote branch

### `branch -D <name>`
同 `-d` 流程，但對未 merge 的 branch 加警告：
```
! feature/login not merged in frontend
  force delete? (y/N)
```

### `branch -m <old> <new>`
1. 掃描哪些 repo 有 `<old>` branch，沒有的給提示跳過
2. 確認清單：
   ```
   Renaming: feature/login → feature/auth

     frontend   has branch → will rename
     backend    has branch → will rename
     shared     no branch  → skip

   Proceed? (y/N)
   ```
3. rename 完成後問是否同步 remote：
   ```
   Sync remote?
     1) yes to all
     2) confirm each
     3) skip
   ```

---

### `stash`
對所有 dirty repo 執行 `git stash`。

### `stash pop`
對所有有 stash 的 repo 執行 `git stash pop`。
衝突時印出衝突檔案 + 提示：
```
repo-a: stash pop conflict
  M src/App.tsx
  → resolve manually, then: git add <file> && git stash drop
```

### `stash apply`
同 `stash pop`，但保留 stash。

### `stash list`
列出所有 repo 的 stash 狀態。

---

### `discard`
自定義指令（無 git 對應），提示後執行 `checkout . && clean -fd`。

---

## 新增指令總結

現有指令沒有、這次新增的：
- `push`
- `fetch`
- `stash` / `stash pop` / `stash apply` / `stash list`
- `branch -d` / `branch -D`
- `branch -m`

---

## Auto-completion 重構

### 現況
`runComplete(prev, cur)` 只看前一個 token，無法處理 subcommand 架構。

### 新架構
改成看整個 command context（所有已輸入的 tokens）：

| 輸入狀態 | 補全內容 |
|----------|----------|
| `gitmulti <TAB>` | 所有 subcommand |
| `gitmulti pull <TAB>` | branch names + `--rebase` |
| `gitmulti push <TAB>` | branch names |
| `gitmulti switch <TAB>` | branch names |
| `gitmulti switch -f <TAB>` | branch names |
| `gitmulti switch -c <TAB>` | （無，新 branch 名自己輸入）|
| `gitmulti branch <TAB>` | `-a`, `--find`, `-d`, `-D`, `-m` |
| `gitmulti branch --find <TAB>` | branch names |
| `gitmulti branch -d <TAB>` | branch names |
| `gitmulti branch -D <TAB>` | branch names |
| `gitmulti branch -m <TAB>` | branch names |
| `gitmulti branch -m <old> <TAB>` | （無，新名稱自己輸入）|
| `gitmulti stash <TAB>` | `pop`, `apply`, `list` |
| `gitmulti -C <TAB>` | repo 目錄名 |
| `gitmulti fetch <TAB>` | （無）|
| `gitmulti status <TAB>` | （無）|
| `gitmulti discard <TAB>` | （無）|

### Branch name 補全格式
動態掃描所有 repo，去重後回傳 branch name 清單（只補 name，不顯示 repo 資訊）。
