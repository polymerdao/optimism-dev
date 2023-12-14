package da

import "github.com/Layr-Labs/eigenda/api/grpc/disperser"

type DAConfig struct {
	Rpc    string
	Client disperser.DisperserClient
}
