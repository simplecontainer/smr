# Branches and Tags

## Branches

The `main` branch represents active development for upcoming releases.  
All new features and fixes should target `main`.

### Currently Maintained Branches

| Branch Name                | Sponsoring Maintainer(s)         | Contribution Status   | Planned End of Maintenance | Known Distributors                     |
|-----------------------------|----------------------------------|-----------------------|-----------------------------|----------------------------------------|
| main (development branch)   | [SMR Maintainers](../MAINTAINERS.md) | N/A                   | -                       | N/A                                    |

---

## Contribution Status

The contribution status helps set clear expectations for what kind of changes a branch will accept:

- **Maintained** → Actively supported by maintainers. Open to contributions, bug fixes, and security patches. Included in security advisories.
- **Maintained (security)** → Only security fixes will be considered. No new features or non-critical bug fixes. Still covered by security advisories.
- **Unmaintained** → No longer supported. Contributions will not be accepted. Not in scope for security advisories.

---

## Tags

Every release of `simplecontainer/smr` is tagged in the repository.  
We follow [Semantic Versioning](https://semver.org) as closely as possible.

Tag format:

```
vX.Y.Z[-suffix[N]]
```

Examples:
- `v1.2.0-smr` → stable release of smr
- `v1.3.0-smrctl` → stable release of smrctl
