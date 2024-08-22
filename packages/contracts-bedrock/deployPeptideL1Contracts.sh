
export OLD_DEPLOY_CONFIG_PATH=./deploy-config/polymer-mainnet.json
export DEPLOY_CONFIG_PATH=./deploy-config/polymer-mainnet-1.json
export IMPL_SALT="polymer-deploy-1"

cat $DEPLOY_CONFIG_PATH

export blockNumber=$(cast to-dec $(cast block --rpc-url $LOCAL_RPC -j | jq -r .number))

echo $blockNumber
# Deploy l2output oracle address
jq --arg new_value $blockNumber '.l2OutputOracleStartingBlockNumber= $new_value' $OLD_DEPLOY_CONFIG_PATH > $DEPLOY_CONFIG_PATH

forge script \
    scripts/Deploy.s.sol:Deploy \
    --sig runPolymerContractsWithStateDump \
    --broadcast \
    --private-key $DUMMY_PRIVATE_KEY \
    --rpc-url $LOCAL_RPC

