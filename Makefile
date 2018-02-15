all: build

################################################################################
##                                 HOME                                       ##
################################################################################
HOME ?= /tmp/gocsi
export HOME


################################################################################
##                                GOPATH                                      ##
################################################################################
# Ensure GOPATH is set and that it contains only a single element.
GOPATH ?= $(HOME)/go
GOPATH := $(word 1,$(subst :, ,$(GOPATH)))
export GOPATH


################################################################################
##                                   DEP                                      ##
################################################################################
DEP ?= ./dep
DEP_DIR := ./vendor/github.com/golang/dep

$(DEP):
	@rm -fr $(DEP_DIR)
	git clone https://github.com/akutz/dep $(DEP_DIR)
	git -C $(DEP_DIR) checkout feature/fetch-gh-pulls
	go build -o $@ $(DEP_DIR)/cmd/dep

ifneq (./dep,$(DEP))
dep: $(DEP)
endif

dep-ensure: | $(DEP)
	$(DEP) ensure -v


########################################################################
##                               CSI SPEC                             ##
########################################################################
CSI_SPEC :=  vendor/github.com/container-storage-interface/spec
CSI_GOSRC := $(CSI_SPEC)/lib/go/csi/csi.pb.go


########################################################################
##                               GOCSI                                ##
########################################################################
CONTEXT_A := context.a
$(CONTEXT_A): context/*.go
	@go install ./$(basename $(@F))
	go build -o "$@" ./$(basename $(@F))

MIDDLEWARE_PKGS := $(addsuffix .a,$(filter-out %.a,$(wildcard middleware/*)))
$(MIDDLEWARE_PKGS): %.a: $(wildcard %/*.go)
	@go install ./middleware/$(basename $(@F))
	go build -o "$@" ./middleware/$(basename $(@F))
middleware: $(MIDDLEWARE_PKGS)
.PHONY: middleware

UTILS_A := utils.a
$(UTILS_A): utils/*.go
	@go install ./$(basename $(@F))
	go build -o "$@" ./$(basename $(@F))

GOCSI_A_PKG_DEPS := $(CONTEXT_A) $(MIDDLEWARE_PKGS) $(UTILS_A)

GOCSI_A := gocsi.a
$(GOCSI_A): $(CSI_GOSRC) *.go $(GOCSI_A_PKG_DEPS)
	@go install .
	go build -o "$@" .


########################################################################
##                               CSI-SP                               ##
########################################################################
CSI_SP_IMPORT := csi-sp
CSI_SP_DIR := $(GOPATH)/src/$(CSI_SP_IMPORT)
CSI_SP := $(CSI_SP_DIR)/csi-sp
CSI_SP_SOCK := $(notdir $(CSI_SP)).sock
CSI_SP_LOG := $(notdir $(CSI_SP)).log
GOCSI_SH_ENV := USE_DEP=true
GOCSI_SH_ENV += DEP_GIT_REFSPECS=+refs/pull/*:refs/pull/origin/*
ifeq (true,$(TRAVIS))
GOCSI_SH_ENV += GOCSI_DEP_SOURCE=https://github.com/$(TRAVIS_REPO_SLUG)
GOCSI_SH_ENV += GOCSI_DEP_REVISION=$(TRAVIS_COMMIT)
endif
$(CSI_SP): | $(DEP)
	DEP=$(abspath $(DEP)) $(GOCSI_SH_ENV) ./gocsi.sh $(CSI_SP_IMPORT)

csi-sp: $(CSI_SP_LOG)
$(CSI_SP_LOG): $(CSI_SP)
	$(MAKE) -C csc
	@echo && \
	  printf '=%.0s' $$(seq 1 80) && printf '\n== ' && \
	  printf "%-74s" "starting $(<F)" && printf ' ==\n' && \
	  printf '=%.0s' $$(seq 1 80) && echo
	CSI_ENDPOINT=$(CSI_SP_SOCK) \
	  X_CSI_LOG_LEVEL=debug \
	  X_CSI_REQ_LOGGING=true \
	  X_CSI_REP_LOGGING=true \
	  X_CSI_SUPPORTED_VERSIONS="0.1.0 0.1.1 0.2.0" \
	  X_CSI_PLUGIN_INFO="My CSI Plug-in,0.1.0,status=online" \
	  $< > $(CSI_SP_LOG) 2>&1 &
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  if grep -q "msg=serving" $(CSI_SP_LOG); then break; \
	  else sleep 0.1; fi \
	done
	@echo && \
	  printf '=%.0s' $$(seq 1 80) && printf '\n== ' && \
	  printf "%-74s" "invoking GetSupportedVersions" && printf ' ==\n' && \
	  printf '=%.0s' $$(seq 1 80) && echo
	csc/csc -e $(CSI_SP_SOCK) i version
	@echo && \
	  printf '=%.0s' $$(seq 1 80) && printf '\n== ' && \
	  printf "%-74s" "invoking GetPluginInfo" && printf ' ==\n' && \
	  printf '=%.0s' $$(seq 1 80) && echo
	csc/csc -e $(CSI_SP_SOCK) i info -v 0.1.0
	@echo && \
	  printf '=%.0s' $$(seq 1 80) && printf '\n== ' && \
	  printf "%-74s" "killing $(<F) with SIGINT" && printf ' ==\n' && \
	  printf '=%.0s' $$(seq 1 80) && echo
	pkill -2 $(<F)
	@echo && \
	  printf '=%.0s' $$(seq 1 80) && printf '\n== ' && \
	  printf "%-74s" "$(<F) log file" && printf ' ==\n' && \
	  printf '=%.0s' $$(seq 1 80) && echo
	@cat $(CSI_SP_LOG)

csi-sp-clean:
	rm -fr $(CSI_SP_LOG) $(CSI_SP_DIR)/*

.PHONY: csi-sp csi-sp-clean


########################################################################
##                               TEST                                 ##
########################################################################
GINKGO := ./ginkgo
GINKGO_PKG := ./vendor/github.com/onsi/ginkgo/ginkgo
GINKGO_SECS := 20
ifeq (true,$(TRAVIS))
GINKGO_SECS := 30
endif
GINKGO_RUN_OPTS := --slowSpecThreshold=$(GINKGO_SECS) -randomizeAllSpecs -p
$(GINKGO): | $(GINKGO_PKG)
	go build -o "$@" "./$|"

ETCD := ./etcd
$(ETCD):
	go get -u -d github.com/coreos/etcd
	go build -o $@ github.com/coreos/etcd

# The test recipe executes the Go tests with the Ginkgo test
# runner. This is the reason for the boolean OR condition
# that is part of the test script. The condition allows for
# the test run to exit with a status set to the value Ginkgo
# uses if it detects programmatic involvement. Please see
# https://goo.gl/CKz4La for more information.
ifneq (true,$(TRAVIS))
test:  build
endif

# Because Travis-CI's containers have limited resources, the Mock SP's
# serial volume access provider's timeout needs to be increased from the
# default value of 0 to 3s. This ensures that lack of system resources
# will not prevent a single, non-concurrent RPC from failing due to an
# OpPending error.
ifeq (true,$(TRAVIS))
export X_CSI_SERIAL_VOL_ACCESS_TIMEOUT=3s
endif
ETCD_ENDPOINT := X_CSI_SERIAL_VOL_ACCESS_ETCD_ENDPOINTS=127.0.0.1:2379

test-serialvolume-etcd: | $(GINKGO) $(ETCD)
	@rm -fr default.etcd etcd.log
	./etcd > etcd.log 2>&1 &
	$(ETCD_ENDPOINT) $(GINKGO) $(GINKGO_RUN_OPTS) \
	  ./middleware/serialvolume/etcd || test "$$?" -eq "197"
	pkill etcd

test-etcd: | $(GINKGO) $(ETCD)
	@rm -fr default.etcd etcd.log
	./etcd > etcd.log 2>&1 &
	$(ETCD_ENDPOINT) $(GINKGO) $(GINKGO_RUN_OPTS) \
	  -skip "Idempotent Create" \
	  ./testing || test "$$?" -eq "197"
	pkill etcd

test-idempotent: | $(GINKGO)
	$(GINKGO) $(GINKGO_RUN_OPTS) -focus "Idempotent Create" \
	  ./testing || test "$$?" -eq "197"

test-utils: | $(GINKGO)
	$(GINKGO) $(GINKGO_RUN_OPTS) ./utils || test "$$?" -eq "197"

test:
	$(MAKE) test-utils
	$(MAKE) test-idempotent
	$(MAKE) test-serialvolume-etcd
	$(MAKE) test-etcd

.PHONY: test \
	    test-utils \
	    test-idempotent \
		test-serialvolume-etcd \
		test-testing-etcd


########################################################################
##                               BUILD                                ##
########################################################################

build: $(GOCSI_A)
	$(MAKE) -C csc $@
	$(MAKE) -C mock $@

clean:
	go clean -i -v . ./csp
	rm -f $(GOCSI_A) $(GOCSI_A_PKG_DEPS)
	$(MAKE) -C csc $@
	$(MAKE) -C mock $@

clobber: clean
	$(MAKE) -C csc $@
	$(MAKE) -C mock $@

.PHONY: build test bench clean clobber
