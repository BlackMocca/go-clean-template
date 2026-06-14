# Generate Go types, server interface and middleware from OpenAPI spec.
# 1. Bundle (dereference all $ref) via redocly → $(SPEC_BUNDLE)
gen-api:
	redocly bundle spec/openapi/openapi.yaml -o spec/openapi/openapi_bundle.yaml
	go generate ./...

gen-interface:
	mockery

test-unit:
	go test $$(find . -name '*_test.go' -exec dirname {} \; |sort -u |grep -vE 'mock|gen|internal/generated') -coverprofile=cover.out
	go tool cover -html=cover.out -o coverage.html
	export unit_total=$$(go test -v $$(find . -name '*_test.go' -exec dirname {} \; |sort -u |grep -vE 'mock|gen|internal/generated')  | grep -c RUN) && echo "Unit Test Total: $$unit_total" && export coverage_total=$$(go tool cover -func cover.out | grep total | awk '{print $$3}') && echo "Coverage Total: $$coverage_total"