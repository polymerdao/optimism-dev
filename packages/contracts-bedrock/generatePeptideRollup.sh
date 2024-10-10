#!/bin/bash

set -euo pipefail
# This script generates a rollup config for only the contracts that peptide needs.
# Must export the following:
# DEPLOY_CONFIG_PATH: the path to the config that was used to deploy (e.g. deploy-config/polymer-staging.json)
# ADDRESSES_PATH: the path where addresses were deployed, generated from running Deploy.s.sol (e.g. "./deployments/11155111-deploy.json")
# ROLLUP_OUT_PATH: the path to output the rollup config path, (e.g. "./rollup.json")
# L1_RPC: the rpc url of the L1 chain (e.g. "https://mainnet.infura.io/v3/your-infura-key")

l1_block_json="$( cast block finalized --rpc-url "$L1_RPC" --json )"
L1BlockNumber="$( printf '%d' "$( jq -r '.number' <<< "$l1_block_json" )" )"
L1BlockHash="$( jq -r '.hash' <<< "$l1_block_json" )"

echo "Fetched L1BlockNumber from RPC : $L1BlockNumber"
echo "Fetched L1BlockHash from RPC:  $L1BlockHash"


# Set these as dummy variables, as they will be overwritten by the peptide init command
export DummyL2BlockNumber=1
export DummyL2BlockHash="0xb7ded8ecf920bd9db1db06540f65718370ca33f813f22113b2fef88c482261ac"
export l2GenesisTime=$(date +%s)

echo "using DummyL2BlockNumber: $DummyL2BlockNumber"
echo "using DummyL2BlockHash: $DummyL2BlockHash"


l1Genesis=$(jq -n --arg blockHash "$L1BlockHash" --arg blockNumber "$L1BlockNumber" '{ "hash": $blockHash, "number": $blockNumber}')
echo "l1Genesis: $l1Genesis"
l2GenesisBlock=$(jq -n --arg blockHash "$DummyL2BlockHash" --arg blockNumber $DummyL2BlockNumber '{ "hash": $blockHash, "number": $blockNumber}')
echo "l1: $l1Genesis: $l2GenesisBlock"

batcherAddr=$(jq '.batchSenderAddress' $DEPLOY_CONFIG_PATH)
gasLimit=30000000
blockTime=$(jq '.l2BlockTime' $DEPLOY_CONFIG_PATH)
maxSequencerDrift=$(jq '.maxSequencerDrift' $DEPLOY_CONFIG_PATH)
sequencerWindowSize=$(jq '.sequencerWindowSize' $DEPLOY_CONFIG_PATH)
channelTimeout=$(jq '.channelTimeout' $DEPLOY_CONFIG_PATH)
l1ChainID=$(($(jq '.l1ChainID' $DEPLOY_CONFIG_PATH)))
l2ChainId=$(jq '.l2ChainID' $DEPLOY_CONFIG_PATH)
regolithTimeHex=$(jq '.l2GenesisRegolithTimeOffset' $DEPLOY_CONFIG_PATH)
regolithTime=null
echo "regolithTimeHex: $regolithTimeHex"
if [[ regolithTimeHex ]]; then
  regolithTimeHex=${regolithTimeHex#0x}
  regolithTime=$((16#$regolithTimeHex))
fi

echo "regolithTime: $regolithTime"

canyonTimeHex=$(jq -r '.l2GenesisCanyonTimeOffset' $DEPLOY_CONFIG_PATH)
canyonTime=null
if [[ canyonTimeHex ]]; then
  canyonTimeHex=${canyonTimeHex#0x}
  canyonTime=$((16#$canyonTimeHex))
fi

deltaTimeHex=$(jq -r '.l2GenesisDeltaTimeOffset' $DEPLOY_CONFIG_PATH)
deltaTime=null
if [[ deltaTimeHex ]]; then
  deltaTimeHex=${deltaTimeHex#0x}
  deltaTime=$((16#$deltaTimeHex))
fi
batchInboxAddress=$(jq '.batchInboxAddress' $DEPLOY_CONFIG_PATH)
depositContractAddress=$(jq '.OptimismPortalProxy' $ADDRESSES_PATH )

system_config=$(jq -n \
  --argjson batcherAddr $batcherAddr \
  --arg overHead "0x0000000000000000000000000000000000000000000000000000000000000000" \
  --arg scalar "0x0000000000000000000000000000000000000000000000000000000000000001" \
  --argjson gasLimit $gasLimit \
  '{ "batcherAddr": $batcherAddr, "overhead": $overHead, "scalar": $scalar, "gasLimit": $gasLimit }')

genesis=$(jq -n \
  --argjson l1 "$l1Genesis" \
  --argjson l2 "$l2GenesisBlock" \
  --argjson l2_time "$l2GenesisTime" \
  --argjson system_config "$system_config" \
  '{ "l1": $l1, "l2": $l2, "l2_time": $l2_time, "system_config": $system_config }' )

rollupConfig=$(jq -n \
  --argjson genesis "$genesis" \
  --argjson block_time "$blockTime" \
  --argjson max_sequencer_drift "$maxSequencerDrift" \
  --argjson seq_window_size "$sequencerWindowSize" \
  --argjson channel_timeout "$channelTimeout" \
  --argjson l1_chain_id "$l1ChainID" \
  --argjson l2_chain_id $l2ChainId \
  --argjson regolith_time "$regolithTime"  \
  --argjson canyon_time "$canyonTime" \
  --argjson delta_time "$deltaTime" \
  --argjson batch_inbox_address "$batchInboxAddress" \
  --argjson deposit_contract_address "$depositContractAddress" \
  --arg l1_system_config_address '0x0000000000000000000000000000000000000000' \
  --arg protocol_versions_address '0x0000000000000000000000000000000000000000' \
  '{ "genesis": $genesis, "block_time": $block_time, "max_sequencer_drift": $max_sequencer_drift, "seq_window_size": $seq_window_size, "channel_timeout": $channel_timeout, "l1_chain_id": $l1_chain_id, "l2_chain_id": $l2_chain_id, "regolith_time": $regolith_time, "canyon_time": $canyon_time, "delta_time":$delta_time, "batch_inbox_address": $batch_inbox_address, "deposit_contract_address": $deposit_contract_address, "l1_system_config_address": $l1_system_config_address, "protocol_versions_address": $protocol_versions_address }'
  )


# # Add plasma config if usePlasma is true
# # TODO: Implement this if we ever use plasma
# # if [[ "$usePlasma" == "true" ]]; then
# #   if [[ -z "$DataAvailabilityChallengeProxy" ]]; then
# #     echo "Error: DataAvailabilityChallengeProxy is not found in deployment but usePlasma is true"
# #     exit 1
# #   fi
# #   rollupConfig="$rollupConfig,\"plasma_config\": {
# #     \"da_challenge_contract_address\": \"$DataAvailabilityChallengeProxy\",
# #     \"da_challenge_window\": 160,
# #     \"da_resolve_window\": 160
# #   }"
# # fi

# # # Write rollupConfig to file
echo "Writing rollup config to" $ROLLUP_OUT_PATH ":"
echo $rollupConfig
echo $rollupConfig > $ROLLUP_OUT_PATH