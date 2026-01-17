# kedge repo remove

## Usage

```
kedge repo remove <name>
```

## Description

Removes a repository from Kedge's database. This stops Kedge from watching the repository, but does **not** stop running containers. To stop containers, use `docker compose down` before removing.

## Arguments

| Argument | Description |
|----------|-------------|
| `name` | Name of the repository to remove |

## Examples

```bash
# Remove a repository
kedge repo remove webapp

# Stop containers first, then remove
cd .kedge/repos/webapp && docker compose down
kedge repo remove webapp
```

## Related Commands

- [kedge repo list](list.md)
- [kedge repo add](add.md)
