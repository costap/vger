---
name: review-renovate-pr
description: Review Renovate dependency update pull requests, verify CI checks and test builds, assess merge safety, and post an assessment comment on GitHub in the user's name. Use whenever the user asks to "review renovate pr", "check renovate prs", "review dependency pr", "is this renovate pr safe to merge", or want automated assessment comments posted on dependency PRs.
---

# Review Renovate PRs

A skill for automated, rigorous review of Renovate dependency update pull requests. It inspects diffs, evaluates version bump risk, validates CI and local builds, and posts a comprehensive assessment comment on GitHub in your name using the GitHub CLI (`gh`).

---

## Workflow Checklist

When asked to review a Renovate PR, follow these sequential steps:

1. **Identify the Target PR**
2. **Inspect Changed Dependencies & Version Bumps**
3. **Verify CI and Test Suite**
4. **Determine Safety Verdict**
5. **Post Assessment Comment on GitHub**
6. **(Optional) Squash and Merge**

---

## Detailed Steps

### Step 1: Identify the Target PR

If the user did not provide a PR number or URL, find open PRs created by Renovate:

```bash
gh pr list --search "author:app/renovate state:open" --json number,title,headRefName,url
```

Select the PR requested or inspect the oldest pending PR first.

---

### Step 2: Inspect Changed Dependencies & Version Bumps

Run the helper script or inspect directly with `gh`:

```bash
# Using the helper script:
.agents/skills/review-renovate-pr/scripts/check-pr.sh <pr-number>

# Or inspect diff directly:
gh pr view <pr-number>
gh pr diff <pr-number>
```

#### Assess Update Classification:
- **Patch (`x.y.Z` ➔ `x.y.Z+1`)**: Bug fixes and minor improvements. Very low risk.
- **Minor (`x.Y.z` ➔ `x.Y+1.0`)**: Backwards-compatible new features. Low-to-moderate risk. Check for newly deprecated APIs.
- **Major (`X.y.z` ➔ `X+1.0.0`)**: Breaking changes likely. High risk. Requires reviewing upstream changelogs and verifying all call sites in `internal/`.
- **GitHub Actions (`actions/*`)**: Review action version tags and permissions.

---

### Step 3: Verify CI and Test Suite

#### 3.1 Check Remote CI Status
Inspect the GitHub Actions check status on the PR:

```bash
gh pr checks <pr-number>
```

#### 3.2 Run Local Verification (Recommended for Non-Trivial Changes)
If CI is pending, has failures, or the update touches core libraries (`google.golang.org/genai`, `firebase/genkit/go`, `charmbracelet/lipgloss`), check out and verify locally:

```bash
# 1. Checkout PR branch
gh pr checkout <pr-number>

# 2. Verify Go module integrity and checksums
go mod verify

# 3. Static analysis & vetting
go vet ./...

# 4. Run test suite
go test -v ./...

# 5. Verify CLI compilation
go build -o /dev/null ./cmd/vger

# 6. Switch back to your working branch
git checkout -
```

---

### Step 4: Determine Safety Verdict

Choose one of three verdicts based on evidence:

| Verdict | Criteria |
| :--- | :--- |
| 🟢 **SAFE TO MERGE** | All CI checks passed, `go mod verify` clean, all unit tests pass, compilation succeeds, and updates are patch/minor with zero breaking API usage in `vger`. |
| 🟡 **ATTENTION REQUIRED** | Major version bump, deprecated API usage detected, or indirect dependencies with notable changes that warrant human eyes. |
| 🔴 **UNSAFE TO MERGE** | CI or unit test failure, compilation error, checksum mismatch, or breaking API changes. |

---

### Step 5: Post Assessment Comment on GitHub

Post a structured review comment on the PR using `gh pr comment`. Because `gh` is authenticated with your GitHub credentials, the comment is posted in your name.

#### Comment Template:

```markdown
### 🤖 Dependency Review Assessment: #<PR_NUMBER>

**Verdict:** <VERDICT_BADGE>

#### 📦 Updated Packages
- `<package-name>`: `<old-version>` ➔ `<new-version>` (<bump-type>)

#### 🔍 Verification Checklist
- [x] **CI Checks**: <Passed / Pending / Failed>
- [x] **Module Verification**: `go mod verify` clean
- [x] **Static Analysis**: `go vet ./...` clean
- [x] **Unit Tests**: All package tests passed
- [x] **Binary Compilation**: `cmd/vger` built successfully

#### 💡 Recommendation
<1-2 sentences summarizing why it is safe to merge or detailing any issues found.>
```

#### Post Command:

```bash
gh pr comment <pr-number> --body "$(cat <<'EOF'
### 🤖 Dependency Review Assessment: #123

**Verdict:** 🟢 **SAFE TO MERGE**

#### 📦 Updated Packages
- `google.golang.org/genai`: `v1.52.1` ➔ `v1.53.0` (minor)

#### 🔍 Verification Checklist
- [x] **CI Checks**: Passed on GitHub Actions
- [x] **Module Verification**: `go mod verify` passed
- [x] **Static Analysis**: `go vet ./...` passed
- [x] **Unit Tests**: All unit tests passed
- [x] **Binary Compilation**: `cmd/vger` builds cleanly

#### 💡 Recommendation
No breaking API changes detected. Unit tests pass and build succeeds cleanly. Safe to merge.
EOF
)"
```

---

### Step 6: Optional Merge Execution

If the verdict is 🟢 **SAFE TO MERGE** and the user requested to merge if safe:

```bash
gh pr merge <pr-number> --squash --delete-branch
```
