# Branches and Tags

> 📌 For details about how patch releases are handled, please see [PATCH-RELEASES.md](PATCH-RELEASES.md).

---

## Branches

The `main` branch represents active development for upcoming releases.  
All new features and fixes should target `main`. Changes to release branches should only happen through **cherry-picked commits**, and only when relevant to maintenance (e.g., bug or security fixes).

Each release branch has a sponsoring maintainer who serves as the point of contact and is responsible for guidance on contributions to that branch.

🚫 **Note:** Release branches are not meant for new feature development. Only bug fixes and security updates are eligible for backports.

### Currently Maintained Branches

| Branch Name                | Sponsoring Maintainer(s)         | Contribution Status   | Planned End of Maintenance | Known Distributors                     |
|-----------------------------|----------------------------------|-----------------------|-----------------------------|----------------------------------------|
| main (development branch)   | [SMR Maintainers](../MAINTAINERS.md) | N/A                   | -                           | N/A                                    |
| 1.2.x                       | [SMR Maintainers](../MAINTAINERS.md) | Maintained            | After 1.3.x release         | [Docker Hub][dockerhub], [GitHub][ghcr] |
| 1.1.x                       | @yourhandle                      | Maintained (security) | 2026-01-15                  | [GitHub][ghcr]                         |
| 1.0.x                       |                                  | Unmaintained          | -                           |                                        |
| Older than 1.0              |                                  | Unmaintained          | -                           |                                        |

[dockerhub]: https://hub.docker.com/r/simplecontainer/smr
[ghcr]: https://github.com/simplecontainer/smr/pkgs/container/smr

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
- `v1.2.0` → stable release
- `v1.3.0-rc1` → first release candidate for `1.3.0`
- `v2.0.0-alpha2` → second alpha for the `2.0.0` release  