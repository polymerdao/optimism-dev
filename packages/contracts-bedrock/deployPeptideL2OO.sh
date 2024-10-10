set -ueo pipefail

this_dir=$(dirname "$0")

export CHAIN_ID=$(cast chain-id --rpc-url "$L1_RPC_URL")

echo "Starting opstack deploy on chainId: $CHAIN_ID and environment $ENVIRONMENT_NAME"

read -p  "Is this correct? hit enter to continue..."
echo "Using L1 RPC: $L1_RPC_URL"
read -p  "Is this correct? hit enter to continue..."

read -p  "Using Impl_SALT: $IMPL_SALT hit enter to continue..."

# Generate deployment config
. $this_dir/scripts/getting-started/config.sh

# Deploy l2output oracle address
echo "this is the config we will use to deploy: "
cat $DEPLOY_CONFIG_PATH

echo "from path $DEPLOY_CONFIG_PATH"
read -p "Does this look correct? Hit enter to continue..."

echo "from deployer address $(cast wallet address --private-key $DEPLOYER_KEY)"
read -p "Does this look correct? Hit enter to continue..."

# Load Safe proxy deployed from above run to use it to deploy L2OOProxy
export CONTRACT_ADDRESSES_PATH="./deployments/$CHAIN_ID-deploy.json"

cd $this_dir
read -p "Reading from contract addresses: "
cat $CONTRACT_ADDRESSES_PATH

# First only deploy contracts related to rollup
forge script \
    scripts/Deploy.s.sol:Deploy \
    --sig runPolymerL2OOContractsWithStateDiff \
    --broadcast \
    --private-key $DEPLOYER_KEY \
    --rpc-url $L1_RPC_URL \
    --slow
