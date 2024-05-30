# Optimism fork


This readme contains information about how we handle the optimism forks

## The repos (optimism v1.7.6 Fjord upgrade and later)

Since the Fjord upgrade, we only need to keep track of a single upstream repo
https://github.com/ethereum-optimism/optimism



## The repos (pre optimism v1.7.6 Fjord upgrade)

We currently keep track of three different repositories

1. The optimism repo, where the "vanilla" op-stack comes from: https://github.com/ethereum-optimism/optimism

2. The Eigen DA repo, which is an optimism fork in itself containing their changes on top of opstack:
   https://github.com/Layr-Labs/optimism

3. Our optimism fork (this repo) which adds our patches on top of vanilla opstack or eigenda's fork.


We are handling the three repos within the same local working copy for simplicity. That means we have three different
"git remotes"


```sh
$ git remote -v
eigenda git@github.com:Layr-Labs/optimism.git (fetch)
eigenda git@github.com:Layr-Labs/optimism.git (push)
optimism        git@github.com:ethereum-optimism/optimism.git (fetch)
optimism        git@github.com:ethereum-optimism/optimism.git (push)
origin  git@github.com:polymerdao/optimism-dev.git (fetch)
origin  git@github.com:polymerdao/optimism-dev.git (push)
```

These remotes where added like so
```sh
$ git remote add eigenda git@github.com:Layr-Labs/optimism.git

$ git remote add optimism git@github.com:ethereum-optimism/optimism.git
```

Also, make sure to pull tags from all remotes

```sh
$ git pull --tags --all
```

If you see any errors when pulling tags, try with `--force`. This is safe as long as there are conflicts with their
tags.

## Our patches

We keep a list of all our patches applied on each version in the `polymer-patches` branch, under the `patches` dir.


## Updating to a new upstream

We always want to keep a straight git history line so keeping up with new upstream versions is easier. That means
we need to rebase our patches on top of new upstream versions. I'll try to explain that process using the latest
update as an example.


### Check the new upstream version

For this example we are going to update our EigenDA enabled fork. So, first check for new versions on the `eigenda`
remote


```sh
$ git pull eigenda
From github.com:Layr-Labs/optimism
...
```

They don't seem to have a clean process to do updates so we have to use common sense and find the right starting
point. Always do updates from release tags, **never** from branches or random commits. So, looking at their tags we
can safely asume their latest version is `v1.7.2-eigenda/v0.6.1-v2`

```sh
$ git tag --list | grep eigenda
1.7.2-eigenda/v0.6.0
1.7.2-eigenda/v0.6.1
v1.4.2-polymer-eigenda-0.3.0
v1.4.2-polymer-eigenda-0.3.1
v1.4.2-polymer-eigenda-0.3.2
v1.4.2-polymer-eigenda-0.6.1
v1.4.3-eigenda
v1.6.1-eigenda
v1.7.2-eigenda-polymer-v1
v1.7.2-eigenda/v0.5.1
v1.7.2-eigenda/v0.6.0
v1.7.2-eigenda/v0.6.1
v1.7.2-eigenda/v0.6.1-v2
```

### Find an upstream tag to fork from

We should only fork from official optimism releases, eg.
https://github.com/ethereum-optimism/optimism/releases/tag/v1.7.6


You can check what's in there with `git log --graph --pretty=reference v1.7.2-eigenda/v0.6.1-v2` or any other
command of your liking. In any case, we want to branch off of that point, so we do something like this

```sh
$ git checkout -b v1.7.2-eigenda-polymer-dev v1.7.2-eigenda/v0.6.1-v2
```

This creates the `v1.7.2-eigenda-polymer-dev` branch starting from the `v1.7.2-eigenda/v0.6.1-v2` tag.


```sh
$ git log -1
commit b10428ac3a2609148db8c3fd16b47a49a13f9411 (HEAD -> v1.7.2-eigenda-polymer-dev, tag: v1.7.2-eigenda/v0.6.1-v2)
...
```

**hacks required to circumvent git hooks**
```sh
# remove the git hooks
rm -rf .git/hooks

# modify `.husky/pre-commit`
git diff | cat
diff --git a/.husky/pre-commit b/.husky/pre-commit
index 91d78423d..9af77529a 100755
--- a/.husky/pre-commit
+++ b/.husky/pre-commit
@@ -1,4 +1,5 @@
 #!/bin/sh
+exit 0
 . "$(dirname "$0")/_/husky.sh"
```

No need to commit the hacks to our fork.

### Cherry pick commits from a previous fork

Now we are ready to apply our patches. We keep a list of the patches that have been created overtime and we should
keep them up to date for the sake of traceability. However, to me is much easier to work with `git cherry-pick` than
`git am`, so we are going to cherry pick these from the previous tag `v1.4.2-polymer-2`. Since we know there's
four patches to be applied, we can check the last four commits in that branch

```sh
# replace 5 with the number of patches you want to apply, from the first commit since the official optimism tag (eg. op-node/v1.7.6) to the last commit in the tagged polymer fork (eg. v1.7.6-polymer-1)
$ git log --oneline --decorate --graph v1.4.2-polymer-2 -5
* d67634c7e (tag: v1.4.2-polymer-2) Github workflow build and publish docker images inkvi <374203+Inkvi@users.noreply.github.com>, 3 months ago
* 518341f3e (tag: v1.4.2-polymer-1.1) adding support for marshalling/unmarshalling blocks jlafiandra6 <jllafiandra96@gmail.com>, 4 months ago
* 489a804bd (tag: v1.4.2-polymer) add roothash to L1BlockInfo and ignore storage proofs Derek <derek@polymerlabs.org>, 4 months ago
* 39f6f7c2f update tests to conform with Polymer Derek <derek@polymerlabs.org>, 4 months ago
```

Now it's just a matter of applying these and working through the conflicts

```sh
$ git cherry-pick 39f6f7c2f 489a804bd 518341f3e d67634c7e
```

Once you do, the branch should look like this

```sh
$ git log --oneline --decorate --graph -5
* 65acd28e1 (HEAD -> v1.7.2-eigenda-polymer-dev) Github workflow build and publish docker images inkvi <374203+Inkvi@users.noreply.github.com>, 3 months ago
* 3ac3b7308 adding support for marshalling/unmarshalling blocks jlafiandra6 <jllafiandra96@gmail.com>, 4 months ago
* e972d1009 add roothash to L1BlockInfo and ignore storage proofs Derek <derek@polymerlabs.org>, 4 months ago
* 46ca869f7 update tests to conform with Polymer Derek <derek@polymerlabs.org>, 4 months ago
* b10428ac3 (tag: v1.7.2-eigenda/v0.6.1-v2) Merge pull request #9 from Layr-Labs/wait-for-finalization Teddy Knox <teddy@eigenlabs.org>, 3 weeks ago
```

That is, our four patches applied on top of `v1.7.2-eigenda/v0.6.1-v2`. We can now tag the repo with something like
`v1.7.2-eigenda-polymer-v1` and push to our `origin` remote.
