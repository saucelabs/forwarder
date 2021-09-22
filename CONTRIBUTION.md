The following guidelines aim to point to a direction that should drive the codebase to increased quality.

- Each package should have a `doc.go` file.
- Think before you make changes, design, then code. Design patterns, and well-established techniques are your friend. They allow to reduce code duplication and complexity and increase reusability and performance.
- Documentation is essential! Relevant comments should be added focusing on the **why**, not in the **what**. _Pay attention to the punctuation and casing patterns_
- Pay attention to how the code is vertically spaced and positioned, also sorted (always ascending) for example, the content of a struct, `vars` and `const` declarations, and etc.
- If you use VSCode IDE, the Go extension is installed, **_and properly setup_**, it should obey the configuration file ([.golangci.yml](.golangci.yml)) for the linter (`golangci`) and show problems the right way, otherwise, just run `$ make lint`. The same thing applies to test. If you open a test file (`*_test.go`), modify and save it, it should automatically run tests and shows coverage; otherwise, just run `$ make test`
- Always run ` $ make coverage lint` before you commit your code; it will save you time!
- If you spotted a problem or something that needs to be modified/improved, do that right way; otherwise, that with `// TODO:`
- Update the [`CHANGELOG.md`](CHANGELOG.md) but not copying your commits messages - that's not its purpose. Use that to plan changes too.
- Don't write tests that test someone else package/APIs, for example, for a function that purely calls some Kubernetes API - they got that already tested, and covered in their own packages. If not, try to change the package you are using for something better. Focus on testing **business logic** and its desired outcomes.