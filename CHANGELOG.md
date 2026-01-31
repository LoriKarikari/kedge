# Changelog

## [0.4.0](https://github.com/LoriKarikari/kedge/compare/v0.3.0...v0.4.0) (2026-01-31)


### Features

* add auth package and database schema for private repo support ([#44](https://github.com/LoriKarikari/kedge/issues/44)) ([c8890a7](https://github.com/LoriKarikari/kedge/commit/c8890a70f9b4654484b58bfdd1628b10fe757bcf))
* add OpenTelemetry metrics and telemetry documentation ([#38](https://github.com/LoriKarikari/kedge/issues/38)) ([137e18a](https://github.com/LoriKarikari/kedge/commit/137e18aad36cb00b9d51c79c8b97c86d7d246937))

## [0.3.0](https://github.com/LoriKarikari/kedge/compare/v0.2.0...v0.3.0) (2026-01-17)


### Features

* add multi-repo support ([#34](https://github.com/LoriKarikari/kedge/issues/34)) ([88c2b15](https://github.com/LoriKarikari/kedge/commit/88c2b15d0c75cbb6ee0ca43b09e117baee7db660))
* **manager:** continue running healthy repos when some fail ([#36](https://github.com/LoriKarikari/kedge/issues/36)) ([eb6beab](https://github.com/LoriKarikari/kedge/commit/eb6beab6c39047ea43cd9e2b3c5a5aaf7c619275))

## [0.2.0](https://github.com/LoriKarikari/kedge/compare/v0.1.0...v0.2.0) (2025-12-29)


### Features

* **config:** add YAML configuration file support ([#31](https://github.com/LoriKarikari/kedge/issues/31)) ([9bf170c](https://github.com/LoriKarikari/kedge/commit/9bf170c35591d642c6851e5f7e16817a142ed9b4))
* **server:** add health and ready endpoints ([#32](https://github.com/LoriKarikari/kedge/issues/32)) ([7e919c0](https://github.com/LoriKarikari/kedge/commit/7e919c098662b919649a2cb871530866fd449577))


### Bug Fixes

* **docker:** connect containers to default network with DNS aliases ([bd6e44d](https://github.com/LoriKarikari/kedge/commit/bd6e44dedcbde759a2a3a0ae86b4dbd3c421c6d1))
* **docker:** connect containers to default network with DNS aliases ([7680af9](https://github.com/LoriKarikari/kedge/commit/7680af9bacd2ac8baa84b82e5a4a5a7ab6768cd3))

## 0.1.0 (2025-12-29)


### Features

* add Dockerfile with distroless image ([#14](https://github.com/LoriKarikari/kedge/issues/14)) ([a7c5669](https://github.com/LoriKarikari/kedge/commit/a7c5669b32a10dc42dd18c88e4ba49148e648f95))
* **cli:** add Cobra CLI with serve, status, diff, sync, history, rollback commands ([#11](https://github.com/LoriKarikari/kedge/issues/11)) ([6eec1d8](https://github.com/LoriKarikari/kedge/commit/6eec1d81547e1b86d75a81f3339b7d45d7ee651a))
* **cli:** add version command ([#15](https://github.com/LoriKarikari/kedge/issues/15)) ([4fe2698](https://github.com/LoriKarikari/kedge/commit/4fe2698dba57ab91aea0ebe293faee286eb54eb1))
* **config:** add config loader with yaml support ([#3](https://github.com/LoriKarikari/kedge/issues/3)) ([c76befd](https://github.com/LoriKarikari/kedge/commit/c76befdf10685df7aefda364ff65111eca8e9f55))
* **controller:** add controller loop for GitOps workflow ([#10](https://github.com/LoriKarikari/kedge/issues/10)) ([55cdb75](https://github.com/LoriKarikari/kedge/commit/55cdb753bf696fe26d1904dcca3a85392975bc8e))
* **controller:** add drift watcher for continuous reconciliation ([#12](https://github.com/LoriKarikari/kedge/issues/12)) ([b4701e0](https://github.com/LoriKarikari/kedge/commit/b4701e0dcffad68302621712c830a26be3901d35))
* **docker:** add docker client ([#6](https://github.com/LoriKarikari/kedge/issues/6)) ([d7d9f7a](https://github.com/LoriKarikari/kedge/commit/d7d9f7a8d9944cb77c32ffd0164751341c91c85b))
* **docker:** add drift detection for services ([#7](https://github.com/LoriKarikari/kedge/issues/7)) ([0769c95](https://github.com/LoriKarikari/kedge/commit/0769c9524fb30f5f46ee163438da22da09a22599))
* **git:** add git watcher with polling support ([#5](https://github.com/LoriKarikari/kedge/issues/5)) ([52281bf](https://github.com/LoriKarikari/kedge/commit/52281bf6e6583f11b844d1de0a905b2075baa5dd))
* **reconcile:** add reconciler with auto/notify/manual modes ([#8](https://github.com/LoriKarikari/kedge/issues/8)) ([abfb45d](https://github.com/LoriKarikari/kedge/commit/abfb45d1a5dc24f6d8c5ae6f85835048a4fb05b7))
* rename project from go-starter to kedge ([#2](https://github.com/LoriKarikari/kedge/issues/2)) ([31fdb16](https://github.com/LoriKarikari/kedge/commit/31fdb16b42d594267e416c3bfec994909b28b3a9))
* **state:** add SQLite store for deployment history ([#9](https://github.com/LoriKarikari/kedge/issues/9)) ([445dfa8](https://github.com/LoriKarikari/kedge/commit/445dfa828bc0bcdeb29b12409f7f7897580b9e81))


### Bug Fixes

* use release-please config files ([#24](https://github.com/LoriKarikari/kedge/issues/24)) ([60f5e7f](https://github.com/LoriKarikari/kedge/commit/60f5e7f3558bf91490c26dd9a969aa7adcd5c39c))


### Reverts

* remove docs accidentally added in [#15](https://github.com/LoriKarikari/kedge/issues/15) ([#16](https://github.com/LoriKarikari/kedge/issues/16)) ([76a52e8](https://github.com/LoriKarikari/kedge/commit/76a52e86d77ba23497baffcdd27c5fd0dc725abb))
