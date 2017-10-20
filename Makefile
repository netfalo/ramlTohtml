PACKAGE  = github.com/netfalo/raml2html
GOPATH   = $(CURDIR)/.gopath
BASE     = $(GOPATH)/src/$(PACKAGE)

$(BASE):
    @mkdir -p $(dir $@)
    @ln -sf $(CURDIR) $@

.PHONY: all
all: | $(BASE)
    cd $(BASE) && $(GO) build -o bin/$(PACKAGE) main.go

GLIDE = glide

glide.lock: glide.yaml | $(BASE)
    cd $(BASE) && $(GLIDE) update
    @touch $@

vendor: glide.lock | $(BASE)
    cd $(BASE) && $(GLIDE) --quiet install
    @ln -sf . vendor/src
    @touch $@
