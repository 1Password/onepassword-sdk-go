## How to Prepare a Release for the Go SDK

Before running this script, the user must make sure that they have the write permissions to the Go SDK repository.

Step 1. Make any changes to the SDK as required on a feature branch or main branch.
NOTE: If ran on a main branch, a release branch will be created.

Step 2. Go to the root of the repo and run 
```
make prep-release
```
Follow the scripts instructions and the release has now been prepped.

Step 3. Ensure that the correct files have been updated - i.e. version/build files, release-notes have been updated. Suggest doing a `git diff` to see the changes.

Step 4. Ensure your GITHUB_TOKEN environment variable is set as this will allow you to create the tags/release and push it.

Step 5. If everything looks good, at the root of the repo, run:
```
make release
```
Step 6. Congratulations, you have released the newest Go SDK!