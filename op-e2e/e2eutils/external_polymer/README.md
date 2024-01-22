# external_polymer shim

See the [external_geth README file](../external_geth/README.md) for context.

## run the shim tests

This test makes sure the shim is able to start the polymer execution engine.
Simply run

```sh
cd op-e2e/external_polymer
make test-shim
```

## run the e2e tests

This is a starting point while polymer as an L2 worked on. This command runs a single test that will
check for basic communication between the op-node and the execution engine

```sh
cd op-e2e
make -C external_polymer/ build
go test --externalL2 ./external_polymer/ -v -run TestSystemE2E
```

To run all e2e tests, we can use the following target

```sh
cd op-e2e
make -C external_polymer/ build
make test-external-polymer
```

**Note that the e2e tests expect to find the shim and polymer within the `external_polymer` directory
and there's no internal hook to automatically build them.**