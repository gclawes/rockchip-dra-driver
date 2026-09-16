# Contributing

PRs against `master` are welcome. CI runs lint, unit tests, and kind e2e
(mock devices) on arm64.

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
| `fix:` / `perf:` | patch (`0.x.y+1`) |
| `docs:` `chore:` `ci:` `test:` `style:` `refactor:` | none |

Common scopes: `npu`, `gpu`, `api`, `helm`, `ci`, `docs`, `discovery`, `deps`.

Until **1.0.0**, do **not** use `feat!` or a `BREAKING CHANGE:` footer; those
would jump the version to 1.0.0. Breaking API changes are still `feat(api): …`
with a note in the body. A human will cut 1.0.0 when the API is stable.

Go module bumps from Dependabot use `fix(deps):` (they ship in the image).
GitHub Actions bumps use `chore(deps):` (CI only, no release).

More project conventions (including for coding agents) are in [AGENTS.md](AGENTS.md).
