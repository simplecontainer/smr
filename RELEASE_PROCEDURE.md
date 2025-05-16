# Release procedure
Commit to main branch empty commit following format:

```bash
git commit -m "Release [patch]" OR
git commit -m "Release [minor]" OR
git commit -m "Release [major]"
```

Afterward run the workflow pipeline and choose which artifacts need to be released.