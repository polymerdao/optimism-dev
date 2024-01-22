build: shim polymer-peptide
.PHONY: build

PEPTIDE_DIR := ../../../polymerase/chain/cmd/peptide

polymer-peptide:
	@test -d $(PEPTIDE_DIR) || { echo "$(PEPTIDE_DIR) does not exist" ; exit 1; }
	cd $(PEPTIDE_DIR) && go build -o $(realpath .)/polymer-peptide ./
.PHONY: polymer-peptide

shim: main.go
	go build -o shim .
.PHONY: shim

test-shim: build
	go test -v ./...
.PHONY: test-shim

clean:
	rm -f shim polymer-peptide
.PHONY: clean

.DEFAULT_GOAL := build