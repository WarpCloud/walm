#!/bin/bash
set -e
export TRUNK=trunk
export SOURCE_BRANCH=dev
export TARGET_BRANCH=master
git clone https://github.com/WarpCloud/walm.git
cd walm
git checkout $TARGET_BRANCH
git remote add $TRUNK ssh://git@172.16.1.41:10022/TDC/WALM.git
git fetch --all
git merge -s ours --no-commit $TRUNK/$SOURCE_BRANCH
for i in `cat .mirror`; do git checkout $TRUNK/$SOURCE_BRANCH -- $i; done
CHANGED=`git status -s | wc -l`
if [[ $CHANGED -gt 0 ]]; then git add . && git commit -m "Sync to mirror" && git push origin $TARGET_BRANCH; fi
echo "SUCCESS"
