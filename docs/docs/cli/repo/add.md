# kedge repo add

## Usage

```
kedge repo add <url> [flags]
```

## Description

Registers a Git repository with Kedge. The repository must contain a `kedge.yaml` configuration file at its root. Once registered, the repository will be cloned and deployed when you run `kedge serve`.

## Flags

| Option | Description | Default |
|--------|-------------|---------|
| `--name` | Custom name for the repository | Derived from URL |
| `--branch` | Branch to watch | `main` |

## Examples

```bash
# Register with defaults (name derived from URL)
kedge repo add https://github.com/acme/webapp

# Register with custom name
kedge repo add https://github.com/acme/webapp --name production

# Register specific branch
kedge repo add https://github.com/acme/webapp --branch develop

# Full example
kedge repo add https://github.com/acme/webapp --name staging --branch release
```

## Related Commands

- [kedge repo list](list.md)
- [kedge repo remove](remove.md)
- [kedge serve](../serve.md)
