# Contributing

PRs against `master` are welcome. CI runs lint, unit tests, and kind e2e
(mock devices) on arm64. A single releasing change merged to `master`
publishes immediately. Batch several of them on a `prep/X.Y.Z` branch
first; see [Release trains](#release-trains).

## Commit messages

Releases are cut by [semantic-release](https://semantic-release.org/) from
[Angular conventional commits](https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#-git-commit-guidelines)
on `master`. The subject must be:

```
type(scope): subject
```

The subject is imperative, lowercase after the type, and ≤72 characters. The
body explains *why*.

`CHANGELOG.md` and the GitHub Release notes are generated from these subjects.
**One changelog-worthy change per commit.** Do not mix unrelated features, Helm,
CI, and API edits in a single commit.

### What cuts a release

| Type | Release (while we are on 0.x) |
|---|---|
| `feat:` | minor (`0.x+1.0`) |
| `fix:` / `perf:` / `revert:` | patch (`0.x.y+1`) |
| `docs:` `chore:` `ci:` `test:` `style:` `refactor:` | none |

Common scopes: `npu`, `gpu`, `api`, `helm`, `ci`, `docs`, `discovery`, `deps`.

Until **1.0.0**, do **not** use `feat!` or a `BREAKING CHANGE:` footer; those
would jump the version to 1.0.0. Breaking API changes are still `feat(api): …`
with a note in the body. A human will cut 1.0.0 when the API is stable.

Go module bumps from Dependabot use `fix(deps):` (they ship in the image).
GitHub Actions bumps use `chore(deps):` (CI only, no release).

More project conventions (including for coding agents) are in [AGENTS.md](AGENTS.md).

## Release trains

semantic-release publishes on every push to `master` (and, later, to an
`N.x` or `N.N.x` maintenance branch). It does not publish from `prep/*`.
Several `feat` or `fix` commits merged to `master` together become **one**
version, the highest bump among them, and each subject stays in
`CHANGELOG.md`. That only works if those commits reach `master` intact.

Use a train when more than one releasing pull request (`feat`, `fix`,
`perf`, `revert`, including Dependabot `fix(deps)`) should share a version.
A single releasing pull request still targets `master`. Docs, chore, and CI
pull requests never need a train.

Do not name the branch `release/*`. Push to `N.x` / `N.N.x` already
publishes. `prep/X.Y.Z` is not one of those branches. The name is the
version you expect, not a channel semantic-release reads.

| Planned commits, from the current tag | Branch |
|---|---|
| any `feat:` | next minor, today `prep/0.7.0` if the tag is `0.6.0` |
| only `fix:` / `perf:` / `revert:` | next patch, today `prep/0.6.1` |
| `feat!` or `BREAKING CHANGE:` | do not open a train |

1. Branch `prep/X.Y.Z` from `origin/master` and push it. Feature and fix
   pulls can target that branch immediately. GitHub will not open a pull
   request until the branch has a diff, so do not open one yet.
2. After the first releasing commit is on the branch, open a **draft**
   pull request into `master`. Use
   [`.github/PULL_REQUEST_TEMPLATE/prep-release.md`](.github/PULL_REQUEST_TEMPLATE/prep-release.md)
   (`gh pr create --draft --body-file` that file, or the template dropdown).
   Title: `chore: release X.Y.Z`. Leave it a draft until the set is
   complete. A draft cannot be merged.
3. Target later feature and fix pulls at `prep/X.Y.Z`, not `master`.
   Squashing each of those into the train is fine. Each should already be
   one conventional commit.
4. Retarget Dependabot pulls that belong in this version:
   `gh pr edit <n> --base prep/X.Y.Z`. Do not set `target-branch` in
   `dependabot.yml`. New Dependabot pulls keep opening against `master`.
   A `fix(deps)` merged there publishes a patch in the middle of the train.
5. While the draft is open, do not merge `feat` / `fix` / `perf` to
   `master`. `docs` / `chore` / `ci` may. After they do, merge `master`
   into the train. Do not rebase the train if other pulls are based on it.
6. CI and the pre-release image run on the train pull request, the same as
   any other pull request. The image tag is `git describe` plus the pull
   request number, not `X.Y.Z`.
7. When the set is complete, confirm the `release-train` check is green
   (the branch name matches the commits). Mark the pull request ready and
   merge it with a **merge commit**. Rebase-merge only if the train is
   linear. Never squash.
8. Delete `prep/X.Y.Z` after the merge. Closing the pull request deletes
   its `*-pr-<n>` preview tags. Deleting the branch deletes
   `*-branch-prep-X-Y-Z` tags.

`master` accepts a merge commit or a rebase, not a squash. One open
`prep/*` pull request per base.

After 1.0, a patch train targets the existing `N.N.x` branch, not `master`.
Do not create that branch until master has moved on: pushing to it
publishes. A train aimed at a maintenance branch is patch-only.
