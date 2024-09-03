#!/bin/bash

set -euo pipefail
# This script generates a rollup config for only the contracts that peptide needs.
# Must export the following:
# DEPLOYMENT_CONFIG_PATH: the path to the config that was used to deploy (e.g. deploy-config/polymer-staging.json)
# ADDRESSES_PATH: the path where addresses were deployed, generated from running Deploy.s.sol (e.g. "./deployments/11155111-deploy.json")
# ROLLUP_OUT_PATH: the path to output the rollup config path, (e.g. "./rollup.json")
# GENESIS_PATH: the path of the peptide genesis generated from running the peptide init command (e.g. "./config/genesis.json")
# L1_RPC: the rpc url of the L1 chain (e.g. "https://mainnet.infura.io/v3/your-infura-key")

export L1BlockNumber=$(cast block-number --rpc-url=$L1_RPC)
export L1BlockHash=$(cast block --rpc-url=$L1_RPC -j | jq .hash)

echo "Fetched L1BlockNumber from RPC : $L1BlockNumber"
echo "Fetched L1BlockHash from RPC:  $L1BlockHash"

export L2BlockNumber=$(jq '.genesis_block.number' $GENESIS_PATH)
export L2BlockHash=$(jq  '.genesis_block.hash' $GENESIS_PATH)
echo "read L2BlockNumber from genesis : $L2BlockNumber"
echo "read L2BlockHash from genesis : $L2BlockHash"

export unformattedGenesisTime=$(jq -r '.genesis_time' $GENESIS_PATH)
export l2GenesisTime=$(date -d $unformattedGenesisTime +%s) # this will clip the milliseconds but print out a warning, which we can ignore

l1Genesis=$(jq -n --argjson blockHash $L1BlockHash --argjson blockNumber $L1BlockNumber '{ "hash": $blockHash, "number": $blockNumber}')
l2GenesisBlock=$(jq -n --argjson blockHash $L2BlockHash --argjson blockNumber $L2BlockNumber '{ "hash": $blockHash, "number": $blockNumber}')

batcherAddr=$(jq '.batchSenderAddress' $DEPLOYMENT_CONFIG_PATH)
gasLimitStr=$(jq -r '.l2GenesisBlockGasLimit' $DEPLOYMENT_CONFIG_PATH)
gasLimitStr=${gasLimitStr#0x}
gasLimit=$((16#$gasLimitStr))
blockTime=$(jq '.l2BlockTime' $DEPLOYMENT_CONFIG_PATH)
maxSequencerDrift=$(jq '.maxSequencerDrift' $DEPLOYMENT_CONFIG_PATH)
sequencerWindowSize=$(jq '.sequencerWindowSize' $DEPLOYMENT_CONFIG_PATH)
channelTimeout=$(jq '.channelTimeout' $DEPLOYMENT_CONFIG_PATH)
l1ChainID=$(($(jq '.l1ChainID' $DEPLOYMENT_CONFIG_PATH)))
l2ChainId=$(jq '.l2ChainID' $DEPLOYMENT_CONFIG_PATH)
regolithTimeHex=$(jq -r '.l2GenesisRegolithTimeOffset' $DEPLOYMENT_CONFIG_PATH)
regolithTime=null
if [[ regolithTimeHex ]]; then
  regolithTimeHex=${regolithTimeHex#0x}
  regolithTime=$((16#$regolithTimeHex))
fi

echo "regolithTime: $regolithTime"

canyonTimeHex=$(jq -r '.l2GenesisCanyonTimeOffset' $DEPLOYMENT_CONFIG_PATH)
canyonTime=null
if [[ canyonTimeHex ]]; then
  canyonTimeHex=${canyonTimeHex#0x}
  canyonTime=$((16#$canyonTimeHex))
fi

deltaTimeHex=$(jq -r '.l2GenesisDeltaTimeOffset' $DEPLOYMENT_CONFIG_PATH)
deltaTime=null
if [[ deltaTimeHex ]]; then
  deltaTimeHex=${deltaTimeHex#0x}
  deltaTime=$((16#$deltaTimeHex))
fi
batchInboxAddress=$(jq '.batchInboxAddress' $DEPLOYMENT_CONFIG_PATH)
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