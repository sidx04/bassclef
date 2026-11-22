# Makefile
#
# Override variables on the command line, e.g.:
#   make bench BENCH='^BenchmarkFromBuffer$$' COUNT=10
#   make bench-save OUT=old.txt
#   make bench-compare OLD=old.txt NEW=new.txt
#   make profile-mem PROFILE_BENCH='^BenchmarkFromBuffer$$' PROFILE_PKG=./internal/fingerprint/

BINARY        ?= bassclef
CMD           := ./cmd/bassclef
BENCH_PKGS    ?= ./internal/dsp/ ./internal/fingerprint/
BENCH         ?= .
COUNT         ?= 1
OUT           ?= new.txt
OLD           ?= old.txt
NEW           ?= new.txt
PKG           ?= ./...
RUN           ?=
PROFILE_BENCH ?= ^BenchmarkFromBuffer$$
PROFILE_PKG   ?= ./internal/fingerprint/

.DEFAULT_GOAL := help

.PHONY: help build run test test-v test-race cover vet fmt tidy check \
        bench bench-save bench-compare profile-cpu profile-mem clean

help: 
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## -- Build -- 

build: ## build the bassclef binary
	go build -o $(BINARY) $(CMD)

run: ## Run the CLI (ARGS='recognize song.wav')
	go run $(CMD) $(ARGS)

##  test 

test: ## Run tests (PKG=./internal/dsp/ RUN=TestName to narrow)
	go test $(if $(RUN),-run '$(RUN)') $(PKG)

test-v: ## Run tests verbosely
	go test -v $(if $(RUN),-run '$(RUN)') $(PKG)

test-race: ## Run tests with the race detector
	go test -race $(PKG)

cover: ## Run tests with coverage and open the HTML report
	go test -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out

##  quality 

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go code
	gofmt -s -w .

tidy: ## Tidy go.mod / go.sum
	go mod tidy

check: vet test ## Run vet and tests

##  benchmarks 

bench: ## Run benchmarks (BENCH=regex COUNT=n BENCH_PKGS=...)
	go test -run '^$$' -bench '$(BENCH)' -benchmem -count=$(COUNT) $(BENCH_PKGS)

bench-save: ## Run benchmarks 10x and save to OUT (default new.txt)
	go test -run '^$$' -bench '$(BENCH)' -benchmem -count=10 $(BENCH_PKGS) | tee $(OUT)

bench-compare: ## Compare OLD and NEW benchmark files with benchstat
	@command -v benchstat >/dev/null 2>&1 || { \
		echo "benchstat not found. Install: go install golang.org/x/perf/cmd/benchstat@latest"; exit 1; }
	benchstat $(OLD) $(NEW)

profile-cpu: ## CPU profile one benchmark (PROFILE_BENCH, PROFILE_PKG)
	go test -run '^$$' -bench '$(PROFILE_BENCH)' -cpuprofile cpu.out $(PROFILE_PKG)
	go tool pprof -top cpu.out

profile-mem: ## Memory profile one benchmark (PROFILE_BENCH, PROFILE_PKG)
	go test -run '^$$' -bench '$(PROFILE_BENCH)' -memprofile mem.out $(PROFILE_PKG)
	go tool pprof -top -sample_index=alloc_space mem.out

##  cleanup 

clean: ## Remove binary, coverage, profiles, and test binaries
	rm -f $(BINARY) coverage.out cpu.out mem.out *.test
