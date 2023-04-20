/*
Package staticlint provides an analysis/multichecker tool with a custom list of analyzers.

Staticlint uses the following analyzer:

  - golang.org/x/tools/go/analysis/passes
  - github.com/dominikh/go-tools/tree/master/staticcheck
  - github.com/gordonklaus/ineffassign
  - github.com/kisielk/errcheck

# Analyzer configuration

The following staticcheck checks are enabled by default:
  - SA
  - SA1014
  - SA2003
  - SA3001
  - SA4006
  - SA5000
  - SA6003
  - SA9005
  - S1003
  - ST1003
  - QF1012

The full list of checks can be found at https://staticcheck.io/docs/checks.

To change the list of staticcheck checks, you need to specify the "conf" flag:

	./staticlint -conf="conf_filepath" ./...

# Static analysis commands

To run a specific analyzer, you must specify the analyzer flag.
For example:

	./staticlint -exitmain -errcheck ./...
	...
	./staticlint -ineffassign -SA1014 ./...
	...
*/
package main
