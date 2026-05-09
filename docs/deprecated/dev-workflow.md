name: Development Workflow

on:
  push:
    branches: [ '*' ]
  pull_request:
    branches: [ '*' ]

# Standard: Pre-commit hook installation for local development
# See .github/workflows/CODING_STANDARDS.md section 4.2
jobs:
  pre-commit-checks:
    name: Pre-commit Checks
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Install pre-commit hook
        run: |
          chmod +x scripts/install-hooks.sh
          bash scripts/install-hooks.sh

      - name: Format check
        run: |
          if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
            echo "The following files are not formatted:"
            gofmt -s -l .
            echo "Run 'go fmt ./...' to fix formatting"
            exit 1
          fi

      - name: Quick lint
        run: |
          go install honnef.co/go/tools/cmd/staticcheck@latest
          staticcheck -checks=all ./...

      - name: Quick test
        run: go test -short -v ./...

  build-check:
    name: Build Check
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Download dependencies
        run: go mod download

      - name: Build
        run: go build -v -o axis-dev cmd/axis/main.go

      - name: Verify build
        run: ./axis-dev --help
