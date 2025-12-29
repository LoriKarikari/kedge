# Changelog

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
