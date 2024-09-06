#!/usr/bin/env bash

# This script is used to generate the getting-started.json configuration file
# used in the Getting Started quickstart guide on the docs site. Avoids the
# need to have the getting-started.json committed to the repo since it's an
# invalid JSON file when not filled in, which is annoying.

reqenv() {
    if [ -z "${!1}" ]; then
        echo "Error: environment variable '$1' is undefined"
        exit 1
    fi
}

# Check required environment variables
reqenv "GS_ADMIN_ADDRESS"
reqenv "GS_BATCHER_ADDRESS"
reqenv "GS_PROPOSER_ADDRESS"
reqenv "GS_CHALLENGER_ADDRESS"
reqenv "L2_CHAIN_ID"
reqenv "L1_RPC_URL"
reqenv "DEPLOY_CONFIG_PATH"
reqenv "BATCH_INBOX_ADDRESS"
reqenv "MAX_SEQUENCER_DRIFT"
reqenv "SEQUENCER_WINDOW_SIZE"
reqenv "CHANNEL_TIMEOUT"
reqenv "USE_FAULT_PROOFS"
reqenv "L2OO_SUBMISSION_INTERVAL"
reqenv "L2OO_STARTING_TIMESTAMP"
reqenv "L2OO_STARTING_BLOCK_NUMBER"
reqenv "FINALIZATION_PERIOD_SECONDS"

# Get the finalized block timestamp and hash
block=$(cast block finalized --rpc-url "$L1_RPC_URL")
l1ChainId=$(cast chain-id --rpc-url "$L1_RPC_URL")
timestamp=$(echo "$block" | awk '/timestamp/ { print $2 }')
blockhash=$(echo "$block" | awk '/hash/ { print $2 }')

# Generate the config file
config=$(cat << EOL
{
  "l1ChainID": $l1ChainId,
  "l2BlockTime": 2,
  "l2ChainID": $L2_CHAIN_ID,
  "l2OutputOracleSubmissionInterval": $L2OO_SUBMISSION_INTERVAL,
  "l2OutputOracleStartingTimestamp": $L2OO_STARTING_TIMESTAMP,
  "l2OutputOracleStartingBlockNumber": $L2OO_STARTING_BLOCK_NUMBER,
  "l2OutputOracleProposer": "$GS_PROPOSER_ADDRESS",
  "l2OutputOracleChallenger": "$GS_CHALLENGER_ADDRESS",
  "finalizationPeriodSeconds": $FINALIZATION_PERIOD_SECONDS,
  "proxyAdminOwner": "$GS_ADMIN_ADDRESS",
  "useFaultProofs": $USE_FAULT_PROOFS,
  "batchInboxAddress": "$BATCH_INBOX_ADDRESS",
  "batchSenderAddress": "$GS_BATCHER_ADDRESS",
  "maxSequencerDrift": $MAX_SEQUENCER_DRIFT,
  "sequencerWindowSize": $SEQUENCER_WINDOW_SIZE,
  "channelTimeout": $CHANNEL_TIMEOUT,
  "l2GenesisBlockGasLimit": "0x1c9c380"
}
EOL
)

echo "$config" > $DEPLOY_CONFIG_PATH
