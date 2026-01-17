<p align="center">
  <h1 align="center">kedge</h1>
  <h2 align="center">GitOps controller for Docker Compose</h2>
</p>

<p align="center">
  <a href="https://lorikarikari.github.io/kedge/"><img src="https://img.shields.io/badge/Documentation-394e79?logo=readthedocs&logoColor=00B9FF" alt="Documentation"></a>
  <a href="https://github.com/LoriKarikari/kedge/releases"><img src="https://img.shields.io/github/v/release/LoriKarikari/kedge" alt="Release"></a>
  <a href="https://github.com/LoriKarikari/kedge/blob/main/LICENSE"><img src="https://img.shields.io/github/license/LoriKarikari/kedge" alt="License"></a>
  <a href="https://github.com/LoriKarikari/kedge/actions"><img src="https://github.com/LoriKarikari/kedge/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://goreportcard.com/report/github.com/LoriKarikari/kedge"><img src="https://goreportcard.com/badge/github.com/LoriKarikari/kedge" alt="Go Report Card"></a>
</p>

## What is Kedge?

Kedge watches your Git repositories and automatically deploys Docker Compose applications. When something drifts from the desired state, it fixes it.

## Features

- **Git-driven deployments**: Push and it deploys
- **Drift detection**: Finds stopped or wrong containers
- **Self-healing**: Automatically fixes drift
- **Multi-repo**: Manage multiple applications
- **Rollback**: Restore any previous deployment

## Documentation

To learn more about Kedge, check the [documentation](https://lorikarikari.github.io/kedge/).

## AI Assistance Disclaimer

AI tools (such as Claude, CodeRabbit, Greptile) were used during development, but all code is reviewed and tested by the maintainers to ensure quality and correctness.
