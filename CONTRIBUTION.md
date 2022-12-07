# Contributing to _Forwarder Proxy_

**Thank you for your interest in _Forwarder Proxy_. Your contributions are highly welcome.**

There are multiple ways of getting involved:

- [Report a bug](#report-a-bug)
- [Suggest a feature](#suggest-a-feature)
- [Contribute code](#contribute-code)
- [Release](#release)

## Report a bug
Reporting bugs is one of the best ways to contribute. Before creating a bug report, please check that an [issue](/issues) reporting the same problem does not already exist. If there is such an issue, you may add your information as a comment.

To report a new bug you should open an issue that summarizes the bug and set the label to "bug".

If you want to provide a fix along with your bug report: That is great! In this case please send us a pull request as described in section [Contribute Code](#contribute-code).

## Suggest a feature
To request a new feature you should open an [issue](../../issues/new) and summarize the desired functionality and its use case. Set the issue label to "feature".

## Contribute code

This is an outline of what the workflow for code contributions looks like

- Check the list of open [issues](../../issues). Either assign an existing issue to yourself, or
create a new one that you would like work on and discuss your ideas and use cases.

It is always best to discuss your plans beforehand, to ensure that your contribution is in line with our goals.

- Create a topic branch from where you want to base your work. This is usually main.
- Open a new pull request, label it `work in progress` and outline what you will be contributing
- When creating a pull request, its description should reference the corresponding issue id.
- Make commits of logical units.
- Write good commit messages (see below).
- Push your changes to a topic branch in your fork of the repository.
- As you push your changes, update the pull request with new infomation and tasks as you complete them
- Project maintainers might comment on your work as you progress
- When you are done, remove the `work in progess` label and ping the maintainers for a review

Thanks for your contributions!

### Coding guidelines

Below are a few guidelines we would like you to follow.
If you need help, please reach out to us by opening an issue.

The following guidelines aim to point to a direction that should drive the codebase to increased quality.

- Each package should have a `doc.go` file.
- Think before you make changes, design, then code. Design patterns, and well-established techniques are your friend. They allow to reduce code duplication and complexity and increase reusability and performance.
- Documentation is essential! Relevant comments should be added focusing on the **why**, not in the **what**. _Pay attention to the punctuation and casing patterns_
- Pay attention to how the code is vertically spaced and positioned, also sorted (always ascending) for example, the content of a struct, `vars` and `const` declarations, and etc.
- If you use VSCode IDE, the Go extension is installed, **_and properly setup_**, it should obey the configuration file ([.golangci.yml](.golangci.yml)) for the linter (`golangci`) and show problems the right way, otherwise, just run `$ make lint`. The same thing applies to test. If you open a test file (`*_test.go`), modify and save it, it should automatically run tests and shows coverage; otherwise, just run `$ make test`
- Always run ` $ make coverage lint` before you commit your code; it will save you time!
- If you spotted a problem or something that needs to be modified/improved, do that right way; otherwise, that with `// TODO:`
- Don't write tests that test someone else package/APIs, for example, for a function that purely calls some Kubernetes API - they got that already tested, and covered in their own packages. If not, try to change the package you are using for something better. Focus on testing **business logic** and its desired outcomes.

## Release

1. create a branch, commit update and push
1. once all test pass and PR is approved, merge
1. make a new release by creating a tag that matches the new Sauce Connect version:
   ```sh
   $ git checkout main
   # fetch latest code
   $ git pull origin main
   $ git tag -a "vX.X.X"
   ```
1. push tag
   ```sh
   $ git push origin main --tags
   ```
1. create a "GitHub release" from tag

**Have fun, and happy hacking!**
