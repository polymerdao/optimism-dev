set -ueo pipefail

export IMPL_SALT="polymer-deploy-2"
export DEPLOY_CONFIG_PATH=${BASE_DEPLOY_CONFIG}-1.json # Note: BASE_DEPLOY_CONFIG should be without the .json extension

export blockNumber=$(cast to-dec $(cast block --rpc-url $RPC_URL -j | jq -r .number))

echo $blockNumber
# Deploy l2output oracle address
jq --arg new_value $blockNumber '.l2OutputOracleStartingBlockNumber= $new_value' ${BASE_DEPLOY_CONFIG}.json > $DEPLOY_CONFIG_PATH

echo "using base deploy config: "
cat $DEPLOY_CONFIG_PATH

forge script \
    scripts/Deploy.s.sol:Deploy \
    --sig runPolymerContractsWithStateDump \
    --broadcast \
    --private-key $DEPLOYER_PRIVATE_KEY \
    --rpc-url $RPC_URL

