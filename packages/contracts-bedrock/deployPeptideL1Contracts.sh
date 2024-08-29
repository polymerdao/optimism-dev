set -ueo pipefail

export IMPL_SALT="polymer-deploy-10"
export DEPLOY_CONFIG_PATH=$(readlink -f ${BASE_DEPLOY_CONFIG}.json) # Note: BASE_DEPLOY_CONFIG should be without the .json extension


export CHAIN_ID=$(cast chain-id --rpc-url $RPC_URL)
echo "Starting opstack deploy on chainId: $CHAIN_ID"
read -p  "Is this correct? hit enter to continue..."
echo "Using RPC: $RPC_URL"
read -p  "Is this correct? hit enter to continue..."

export blockNumber=$(cast to-dec $(cast block --rpc-url $RPC_URL -j | jq -r .number))

echo "Using block l2OO start block number $blockNumber "
read -p  "Is this correct? hit enter to continue..."

# Deploy l2output oracle address

echo "this is the config we will use to deploy: "
cat $DEPLOY_CONFIG_PATH

echo "from path $DEPLOY_CONFIG_PATH"
read -p "Does this look correct? Hit enter to continue..."

echo "from deployer address $(cast wallet address --private-key $DEPLOYER_PRIVATE_KEY)"
read -p "Does this look correct? Hit enter to continue..."

forge script \
    scripts/Deploy.s.sol:Deploy \
    --sig runPolymerContractsWithStateDump \
    --broadcast \
    --private-key $DEPLOYER_PRIVATE_KEY \
    --rpc-url $RPC_URL \
    --slow

