# Changelog

All notable changes to the Helmlet project will be documented in this file.

## [v0.0.16] - 2025-03-12
### :bug: Bug Fixes
- [`c421554`](https://github.com/rogosprojects/helmlet/commit/c4215543edfcbc2932a0e3b9407dd8370c3cebba) - update asset name format in release workflow *(commit by [@rogosprojects](https://github.com/rogosprojects))*

### :wrench: Chores
- [`9f89caa`](https://github.com/rogosprojects/helmlet/commit/9f89caafcecb949f2080c5243538ee9fe2eeba45) - update readme.md *(commit by [@rogosprojects](https://github.com/rogosprojects))*


## [Unreleased]
- feat: add versioning to Go build and copy binary to /usr/local/bin/helmlet
- feat: add Dockerfile for building and running Go application
- feat: "strict" option, do not allow missing keys

## [0.0.13] - 2024-10-17
- fix: not-compliant Helm behaviour

## [0.0.12] - 2024-10-17
- feat: indent

## [0.0.11] - 2024-10-17
- feat (experimental): support for non UTF8 charsets
- update tests

## [0.0.10] - 2024-10-14
- support for FilesGet func
- Refactor documentation in README.md

## [0.0.9] - 2024-10-11
- print and comment Helmlet version
- Refactor tar command in builder.sh

## [0.0.8] - 2024-10-11
- feat: value file can be optional
- support for custom delimiters

## [0.0.7] - 2024-10-10
- Refactor file iteration in .gitlab-ci.yml

## [0.0.6] - 2024-10-10
- add version number in binary
- Refactor command line arguments in README.md

## [v0.0.3] - 2024-10-10
- Refactor GitLab CI configuration
- init pipeline
- fix: support for nested keys
- support for inline params; add tests
[v0.0.16]: https://github.com/rogosprojects/helmlet/compare/v0.0.15...v0.0.16
