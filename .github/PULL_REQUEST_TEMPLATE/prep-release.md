<!-- Upcoming release. Merging this pull request is what semantic-release publishes. Do not squash. -->

## Upcoming release

- Anticipated version:
- Base: `master`
- Last published tag:

Merging this pull request to the base branch is what cuts the release. Use **Create a merge commit**. Do not squash. A squash replaces every subject with one message, so `CHANGELOG.md` loses the features. A squash titled `chore:` publishes nothing.

Rebase-merge is acceptable only when the train is a straight line of conventional commits.

The pre-release image for this pull request is tagged from `git describe` plus this pull request number. It is not the anticipated version. That tag exists only after this merges and semantic-release runs.

## Included

-

## Not included

-

## Checklist

- [ ] Branch name is `prep/<version>` and `hack/next-release-version.sh` prints that version for this base
- [ ] Dependabot `fix(deps)` pulls that belong in this version are retargeted here (`gh pr edit <n> --base prep/<version>`)
- [ ] No other `feat` / `fix` / `perf` pull request is about to merge to the base ahead of this one
- [ ] Pre-release image from this pull request was tested, if this train needs a cluster check
- [ ] This pull request is no longer a draft
- [ ] Merge method is a merge commit, or a rebase of a linear train
