.PHONY: docs docs-serve

## docs: generate docs/api/API.md via gomarkdoc
docs:
	go generate ./...

## docs-serve: browse package docs locally via pkgsite
docs-serve:
	pkgsite -open .
