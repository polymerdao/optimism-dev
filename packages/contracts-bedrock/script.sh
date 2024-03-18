#!/usr/bin/env bash

set -xeo pipefail


RPC_URL='http://127.0.0.1:8545'
PRIVATE_KEY='0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80'
# ACCOUNT='0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266'
CONTRACT='src/dispute/DisputeGameFactory.sol'


create_game() {
  GAME_TYPE="$1"
  ROOT_CLAIM="$2"
  EXTRA_DATA="$3"

  cast send \
    --rpc-url "$RPC_URL" \
    --private-key "$PRIVATE_KEY" \
    "$CONTRACT_ADDR" 'create(uint32,bytes32,bytes)()' \
    "$GAME_TYPE" "$ROOT_CLAIM" "$EXTRA_DATA"
}

get_count() {
  cast call --rpc-url http://127.0.0.1:8545 "$CONTRACT_ADDR" 'gameCount()(uint256)'
}

proof() {
  # get full storage layout with `cast storage --rpc-url http://127.0.0.1:8545 $CONTRACT_ADDR`
  #
  # | Name             | Type                                       | Slot | Offset | Bytes | Value | Hex Value                                                          | Contract                                              |
  # |------------------|--------------------------------------------|------|--------|-------|-------|--------------------------------------------------------------------|-------------------------------------------------------|
  # | _disputeGames    | mapping(Hash => GameId)                    | 103  | 0      | 32    | 0     | 0x0000000000000000000000000000000000000000000000000000000000000000 | src/dispute/DisputeGameFactory.sol:DisputeGameFactory |
  # | _disputeGameList | GameId[]                                   | 104  | 0      | 32    | 2     | 0x0000000000000000000000000000000000000000000000000000000000000002 | src/dispute/DisputeGameFactory.sol:DisputeGameFactory |
  #

  GAME_TYPE="$1"
  ROOT_CLAIM="$2"
  EXTRA_DATA="$3"

  KEY="$( cast k "$(  cast abi-encode 'foo(uint32,bytes32,bytes)' "$GAME_TYPE" "$ROOT_CLAIM" "$EXTRA_DATA" )" )"
  SLOT_INDEX="$( cast index bytes32 "$KEY" 103 )"
  cast proof --rpc-url "$RPC_URL" "$CONTRACT_ADDR" "$SLOT_INDEX" | jq
}

get_game() {
  INDEX="$1"
  cast call --rpc-url http://127.0.0.1:8545 "$CONTRACT_ADDR" 'gameAtIndex(uint256)(uint32,uint256,uint256)' "$INDEX"
}

check_contract_addr() {
  if [ -z "$CONTRACT_ADDR" ] ; then
    echo "Set the \$CONTRACT_ADDR env variable"
    exit 1
  fi
}

case "$1" in
  deploy)
    forge create --rpc-url "$RPC_URL" --private-key "$PRIVATE_KEY" "$CONTRACT:DisputeGameFactory"
    ;;

  get-count)
    shift
    check_contract_addr
    get_count "$@"
    ;;

  create-game)
    shift
    check_contract_addr
    create_game "$@"
    ;;

  get-game)
    shift
    check_contract_addr
    get_game "$@"
    ;;


  get-proof)
    shift
    check_contract_addr
    proof "$@"
    ;;

  *)
    echo "unknown command $1"
    exit 1
  ;;
esac
