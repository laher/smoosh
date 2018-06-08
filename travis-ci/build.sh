#!/bin/bash

set -e

# Requires Travis CI env varialbes:
# > TRAVIS_BUILD_DIR
# > TRAVIS_PULL_REQUEST
# > TRAVIS_PULL_REQUEST_BRANCH
# > TRAVIS_BRANCH
# > TRAVIS_COMMIT_RANGE

echo "###########################"
echo "#          BUILD          #"
echo "###########################"

function printDetails() {
    if [ "${TRAVIS_PULL_REQUEST}" != "false" ] ; then
        echo " - PR: ${TRAVIS_PULL_REQUEST}"
        echo " - Base: ${TRAVIS_BRANCH}"
        echo " - Branch: ${TRAVIS_PULL_REQUEST_BRANCH}"
        echo " - Commit: ${TRAVIS_COMMIT_RANGE}"
    else
        echo " - Branch: ${TRAVIS_BRANCH}"
        echo " - Commit: ${TRAVIS_COMMIT_RANGE}"
    fi

    echo ""

    # stop here if there are no difference found in this commit
    if [ "$1" = "" ] ; then
        echo "no directories here affected in this commit"
        exit 0;
    fi

    # print directories to test
    echo "Changed paths:"
    for var in "$@"
    do
        echo " > $var"
    done

    echo ""
}

# for a branch: get all changed dirs between master and branch, which is HEAD
# this accounts for rebases
function getChangedPathsForBranch() {
    git remote set-branches --add origin master
    git fetch
    git diff --name-only origin/master...HEAD | xargs dirname | sort -u
}

# for pull requests and master: use the commit range to find all relevant changes
function getChangedPathsForCommitRange() {
    git diff --name-only "${TRAVIS_COMMIT_RANGE}" | xargs dirname | sort -u
}

# decide what to test based on the PR / branch / master
function getChangedPaths() {
    if [ "${TRAVIS_PULL_REQUEST}" != "false" ] || [ "${TRAVIS_BRANCH}" == "master" ]; then
        getChangedPathsForCommitRange
    else
        getChangedPathsForBranch
    fi
}

cd "${TRAVIS_BUILD_DIR}"
dirs_to_test=$(getChangedPaths)
printDetails ${dirs_to_test}

go test ${dirs_to_test}
