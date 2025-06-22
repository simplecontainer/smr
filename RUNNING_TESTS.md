# Testing

## Local e2e testing
```
make test-e2e IMAGE=smr TAG=$(git rev-parse --short HEAD) TEST_DIR=<DIR_HERE> HOME=/tmp TIMEOUT=200s
```