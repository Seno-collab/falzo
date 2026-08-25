---
name: format-code
description: Format Falzo source code after creating or editing Go, TypeScript, TSX, CSS, or SVG files. Use for coding tasks that modify application source in this repository; do not invoke for read-only work or generated, dependency, cache, migration, and skill files.
---

# Format Code

Before handing off a coding change, format the source files changed by the current task. Do not broaden the diff to pre-existing user changes unless the user explicitly requests repository-wide formatting.

## Formatters

- Run `gofmt -w -- <files>` for changed `.go` files under `be/cmd` or `be/internal`.
- Run Prettier 3.6.2 for changed `.ts`, `.tsx`, and `.css` files under `fe/src`.
- Format changed `.svg` files under `fe/src` with Prettier 3.6.2 using `--parser html`.
- Prefer the repository-local Prettier binary when available. Otherwise use `pnpm dlx prettier@3.6.2` and follow the environment's network-approval requirements.
- Skip generated output, dependencies, caches, migrations, and files outside the task's scope.

Run formatting before the task's final tests so validation covers the final file contents. Confirm the formatted files no longer produce formatter diffs, run checks proportional to the code change, and report the formatter and validation results in the final response.
