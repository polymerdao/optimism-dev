// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const PreimageOracleStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/cannon/PreimageOracle.sol:PreimageOracle\",\"label\":\"preimageLengths\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_mapping(t_bytes32,t_uint256)\"},{\"astId\":1001,\"contract\":\"src/cannon/PreimageOracle.sol:PreimageOracle\",\"label\":\"preimageParts\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_mapping(t_bytes32,t_mapping(t_uint256,t_bytes32))\"},{\"astId\":1002,\"contract\":\"src/cannon/PreimageOracle.sol:PreimageOracle\",\"label\":\"preimagePartOk\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_mapping(t_bytes32,t_mapping(t_uint256,t_bool))\"},{\"astId\":1003,\"contract\":\"src/cannon/PreimageOracle.sol:PreimageOracle\",\"label\":\"zeroHashes\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_array(t_bytes32)16_storage\"},{\"astId\":1004,\"contract\":\"src/cannon/PreimageOracle.sol:PreimageOracle\",\"label\":\"proposalBranches\",\"offset\":0,\"slot\":\"19\",\"type\":\"t_mapping(t_address,t_mapping(t_uint256,t_array(t_bytes32)16_storage))\"},{\"astId\":1005,\"contract\":\"src/cannon/PreimageOracle.sol:PreimageOracle\",\"label\":\"proposalMetadata\",\"offset\":0,\"slot\":\"20\",\"type\":\"t_mapping(t_address,t_mapping(t_uint256,t_userDefinedValueType(LPPMetaData)1007))\"},{\"astId\":1006,\"contract\":\"src/cannon/PreimageOracle.sol:PreimageOracle\",\"label\":\"proposalParts\",\"offset\":0,\"slot\":\"21\",\"type\":\"t_mapping(t_address,t_mapping(t_uint256,t_bytes32))\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_array(t_bytes32)16_storage\":{\"encoding\":\"inplace\",\"label\":\"bytes32[16]\",\"numberOfBytes\":\"512\",\"base\":\"t_bytes32\"},\"t_bool\":{\"encoding\":\"inplace\",\"label\":\"bool\",\"numberOfBytes\":\"1\"},\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_mapping(t_address,t_mapping(t_uint256,t_array(t_bytes32)16_storage))\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e mapping(uint256 =\u003e bytes32[16]))\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_mapping(t_uint256,t_array(t_bytes32)16_storage)\"},\"t_mapping(t_address,t_mapping(t_uint256,t_bytes32))\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e mapping(uint256 =\u003e bytes32))\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_mapping(t_uint256,t_bytes32)\"},\"t_mapping(t_address,t_mapping(t_uint256,t_userDefinedValueType(LPPMetaData)1007))\":{\"encoding\":\"mapping\",\"label\":\"mapping(address =\u003e mapping(uint256 =\u003e LPPMetaData))\",\"numberOfBytes\":\"32\",\"key\":\"t_address\",\"value\":\"t_mapping(t_uint256,t_userDefinedValueType(LPPMetaData)1007)\"},\"t_mapping(t_bytes32,t_mapping(t_uint256,t_bool))\":{\"encoding\":\"mapping\",\"label\":\"mapping(bytes32 =\u003e mapping(uint256 =\u003e bool))\",\"numberOfBytes\":\"32\",\"key\":\"t_bytes32\",\"value\":\"t_mapping(t_uint256,t_bool)\"},\"t_mapping(t_bytes32,t_mapping(t_uint256,t_bytes32))\":{\"encoding\":\"mapping\",\"label\":\"mapping(bytes32 =\u003e mapping(uint256 =\u003e bytes32))\",\"numberOfBytes\":\"32\",\"key\":\"t_bytes32\",\"value\":\"t_mapping(t_uint256,t_bytes32)\"},\"t_mapping(t_bytes32,t_uint256)\":{\"encoding\":\"mapping\",\"label\":\"mapping(bytes32 =\u003e uint256)\",\"numberOfBytes\":\"32\",\"key\":\"t_bytes32\",\"value\":\"t_uint256\"},\"t_mapping(t_uint256,t_array(t_bytes32)16_storage)\":{\"encoding\":\"mapping\",\"label\":\"mapping(uint256 =\u003e bytes32[16])\",\"numberOfBytes\":\"32\",\"key\":\"t_uint256\",\"value\":\"t_array(t_bytes32)16_storage\"},\"t_mapping(t_uint256,t_bool)\":{\"encoding\":\"mapping\",\"label\":\"mapping(uint256 =\u003e bool)\",\"numberOfBytes\":\"32\",\"key\":\"t_uint256\",\"value\":\"t_bool\"},\"t_mapping(t_uint256,t_bytes32)\":{\"encoding\":\"mapping\",\"label\":\"mapping(uint256 =\u003e bytes32)\",\"numberOfBytes\":\"32\",\"key\":\"t_uint256\",\"value\":\"t_bytes32\"},\"t_mapping(t_uint256,t_userDefinedValueType(LPPMetaData)1007)\":{\"encoding\":\"mapping\",\"label\":\"mapping(uint256 =\u003e LPPMetaData)\",\"numberOfBytes\":\"32\",\"key\":\"t_uint256\",\"value\":\"t_userDefinedValueType(LPPMetaData)1007\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"},\"t_userDefinedValueType(LPPMetaData)1007\":{\"encoding\":\"inplace\",\"label\":\"LPPMetaData\",\"numberOfBytes\":\"32\"}}}"

var PreimageOracleStorageLayout = new(solc.StorageLayout)

var PreimageOracleDeployedBin = "0x608060405234801561001057600080fd5b50600436106101515760003560e01c80639f99ef82116100cd578063e03110e111610081578063ec5efcbc11610066578063ec5efcbc14610304578063faf37bc714610317578063fef2b4ed1461032a57600080fd5b8063e03110e1146102c9578063e1592611146102f157600080fd5b8063b4801e61116100b2578063b4801e6114610299578063c3a079ed146102ac578063d18534b5146102b657600080fd5b80639f99ef821461025b578063b2e67ba81461026e57600080fd5b806352f0f3ad116101245780636551927b116101095780636551927b146101df5780637ac547671461020a5780638542cf501461021d57600080fd5b806352f0f3ad146101a157806361238bde146101b457600080fd5b80630359a563146101565780632055b36b1461017c5780633909af5c146101845780634d52b4c914610199575b600080fd5b61016961016436600461224e565b61034a565b6040519081526020015b60405180910390f35b610169601081565b610197610192366004612449565b610482565b005b6101696106d0565b6101696101af366004612535565b6106eb565b6101696101c2366004612570565b600160209081526000928352604080842090915290825290205481565b6101696101ed36600461224e565b601460209081526000928352604080842090915290825290205481565b610169610218366004612592565b6107c0565b61024b61022b366004612570565b600260209081526000928352604080842090915290825290205460ff1681565b6040519015158152602001610173565b6101976102693660046125ed565b6107d7565b61016961027c36600461224e565b601560209081526000928352604080842090915290825290205481565b6101696102a736600461267e565b610cc5565b6101696201518081565b6101976102c4366004612449565b610cf7565b6102dc6102d7366004612570565b6110a5565b60408051928352602083019190915201610173565b6101976102ff3660046126b1565b611196565b6101976103123660046126fd565b61129f565b610197610325366004612796565b611419565b610169610338366004612592565b60006020819052908152604090205481565b73ffffffffffffffffffffffffffffffffffffffff82166000908152601460209081526040808320848452909152812054819061038d9060601c63ffffffff1690565b63ffffffff16905060005b601081101561047a57816001166001036104205773ffffffffffffffffffffffffffffffffffffffff85166000908152601360209081526040808320878452909152902081601081106103ed576103ed6127d2565b01546040805160208101929092528101849052606001604051602081830303815290604052805190602001209250610461565b8260038260108110610434576104346127d2565b01546040805160208101939093528201526060016040516020818303038152906040528051906020012092505b60019190911c908061047281612830565b915050610398565b505092915050565b600061048e8a8a61034a565b90506104b186868360208b01356104ac6104a78d612868565b6114e2565b611522565b80156104cf57506104cf83838360208801356104ac6104a78a612868565b610505576040517f09bde33900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b86604001358860405160200161051b9190612937565b6040516020818303038152906040528051906020012014610568576040517f1968a90200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b83602001358760200135600161057e9190612975565b146105b5576040517f9a3b119900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6105fd886105c3868061298d565b8080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061158392505050565b610606886116de565b83604001358860405160200161061c9190612937565b6040516020818303038152906040528051906020012003610669576040517f9843145b00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5050505073ffffffffffffffffffffffffffffffffffffffff9590951660009081526014602090815260408083209683529590529390932080547fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000016600117905550505050565b60016106de60106002612b14565b6106e89190612b20565b81565b60006106f78686611f7a565b9050610704836008612975565b8211806107115750602083115b15610748576040517ffe25498700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6000602081815260c085901b82526008959095528251828252600286526040808320858452875280832080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660019081179091558484528752808320948352938652838220558181529384905292205592915050565b600381601081106107d057600080fd5b0154905081565b606081156107f0576107e98686612027565b905061082a565b85858080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152509293505050505b3360009081526013602090815260408083208a845290915280822081516102008101928390529160109082845b8154815260200190600101908083116108575750503360009081526014602090815260408083208f845290915290205493945061089992508391506120b09050565b63ffffffff166000036108d8576040517f87138d5c00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6108e28160c01c90565b67ffffffffffffffff1615610923576040517f475a253500000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60006109358260a01c63ffffffff1690565b67ffffffffffffffff16905060006109538360401c63ffffffff1690565b63ffffffff169050600882108015610969575080155b156109f05760006109808460801c63ffffffff1690565b905060008160c01b6000528b356008528351905080601560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008f8152602001908152602001600020819055505050610aa5565b60088210158015610a0e575080610a08600884612b20565b92508210155b8015610a225750610a1f8982612975565b82105b15610aa5576000610a338284612b20565b905089610a41826020612975565b10158015610a4d575086155b15610a84576040517ffe25498700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b3360009081526015602090815260408083208f84529091529020908b013590555b6000610ab78460601c63ffffffff1690565b63ffffffff169050855160208701608882048a1415608883061715610ae4576307b1daf16000526004601cfd5b60405160c8810160405260005b83811015610b94578083018051835260208101516020840152604081015160408401526060810151606084015260808101516080840152508460888301526088810460051b8d013560a883015260c882206001860195508560005b610200811015610b89576001821615610b695782818d0152610b89565b8b81015160009081526020938452604090209260019290921c9101610b4c565b505050608801610af1565b50505050600160106002610ba89190612b14565b610bb29190612b20565b811115610beb576040517f6229572300000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b3360009081526013602090815260408083208f84529091529020610c119086601061219b565b50610c71610c1f838c612975565b60401b7fffffffffffffffffffffffffffffffffffffffff00000000ffffffffffffffff606084901b167fffffffffffffffffffffffffffffffff0000000000000000ffffffffffffffff8716171790565b93508615610c9c5777ffffffffffffffffffffffffffffffffffffffffffffffff84164260c01b1793505b50503360009081526014602090815260408083209c83529b905299909920555050505050505050565b60136020528260005260406000206020528160005260406000208160108110610ced57600080fd5b0154925083915050565b73ffffffffffffffffffffffffffffffffffffffff891660009081526014602090815260408083208b845290915290205467ffffffffffffffff811615610d6a576040517fc334f06900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b62015180610d788260c01c90565b610d8c9067ffffffffffffffff1642612b20565b11610dc3576040517f55d4cbf900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6000610dcf8b8b61034a565b9050610de887878360208c01356104ac6104a78e612868565b8015610e065750610e0684848360208901356104ac6104a78b612868565b610e3c576040517f09bde33900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b876040013589604051602001610e529190612937565b6040516020818303038152906040528051906020012014610e9f576040517f1968a90200000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b846020013588602001356001610eb59190612975565b141580610ee757506001610ecf8360601c63ffffffff1690565b610ed99190612b37565b63ffffffff16856020013514155b15610f1e576040517f9a3b119900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6000610f308360801c63ffffffff1690565b63ffffffff16905080610f498460401c63ffffffff1690565b63ffffffff1614610f86576040517f7b1dafd100000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b610f948a6105c3888061298d565b610f9d8a6116de565b6000610fa88b6120bc565b90506000610fbc8560a01c63ffffffff1690565b67ffffffffffffffff169050600160026000848152602001908152602001600020600083815260200190815260200160002060006101000a81548160ff021916908315150217905550601560008f73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008e8152602001908152602001600020546001600084815260200190815260200160002060008381526020019081526020016000208190555082600080848152602001908152602001600020819055505050505050505050505050505050565b6000828152600260209081526040808320848452909152812054819060ff1661112e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601460248201527f7072652d696d616765206d757374206578697374000000000000000000000000604482015260640160405180910390fd5b506000838152602081815260409091205461114a816008612975565b611155856020612975565b106111735783611166826008612975565b6111709190612b20565b91505b506000938452600160209081526040808620948652939052919092205492909150565b604435600080600883018611156111b55763fe2549876000526004601cfd5b60c083901b6080526088838682378087017ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80151908490207effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f02000000000000000000000000000000000000000000000000000000000000001760008181526002602090815260408083208b8452825280832080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0016600190811790915584845282528083209a83529981528982209390935590815290819052959095209190915550505050565b60006112ab868661034a565b90506112c483838360208801356104ac6104a78a612868565b6112fa576040517f09bde33900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b602084013515611336576040517f9a3b119900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b61133e6121d9565b61134c816105c3878061298d565b611355816116de565b84604001358160405160200161136b9190612937565b60405160208183030381529060405280519060200120036113b8576040517f9843145b00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5050505073ffffffffffffffffffffffffffffffffffffffff9290921660009081526014602090815260408083209383529290522080547fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000016600117905550565b611424816008612b5c565b63ffffffff168263ffffffff1610611468576040517ffe25498700000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b3360009081526014602090815260408083209583529490529290922080547fffffffffffffffff0000000000000000ffffffffffffffffffffffffffffffff1660a09290921b7fffffffffffffffffffffffff00000000ffffffffffffffffffffffffffffffff169190911760809290921b919091179055565b600081600001518260200151836040015160405160200161150593929190612b84565b604051602081830303815290604052805190602001209050919050565b60008160005b6010811015611576578060051b880135600186831c166001811461155b576000848152602083905260409020935061156c565b600082815260208590526040902093505b5050600101611528565b5090931495945050505050565b608881511461159157600080fd5b6020810160208301611612565b8260031b8201518060001a8160011a60081b178160021a60101b8260031a60181b17178160041a60201b8260051a60281b178260061a60301b8360071a60381b171717905061160c816115f7868560059190911b015190565b1867ffffffffffffffff16600586901b840152565b50505050565b61161e6000838361159e565b61162a6001838361159e565b6116366002838361159e565b6116426003838361159e565b61164e6004838361159e565b61165a6005838361159e565b6116666006838361159e565b6116726007838361159e565b61167e6008838361159e565b61168a6009838361159e565b611696600a838361159e565b6116a2600b838361159e565b6116ae600c838361159e565b6116ba600d838361159e565b6116c6600e838361159e565b6116d2600f838361159e565b61160c6010838361159e565b6040805178010000000000008082800000000000808a8000000080008000602082015279808b00000000800000018000000080008081800000000000800991810191909152788a00000000000000880000000080008009000000008000000a60608201527b8000808b800000000000008b8000000000008089800000000000800360808201527f80000000000080028000000000000080000000000000800a800000008000000a60a08201527f800000008000808180000000000080800000000080000001800000008000800860c082015260009060e00160405160208183030381529060405290506020820160208201611e5a565b6102808101516101e082015161014083015160a0840151845118189118186102a082015161020083015161016084015160c0850151602086015118189118186102c083015161022084015161018085015160e0860151604087015118189118186102e08401516102408501516101a0860151610100870151606088015118189118186103008501516102608601516101c0870151610120880151608089015118189118188084603f1c6118918660011b67ffffffffffffffff1690565b18188584603f1c6118ac8660011b67ffffffffffffffff1690565b18188584603f1c6118c78660011b67ffffffffffffffff1690565b181895508483603f1c6118e48560011b67ffffffffffffffff1690565b181894508387603f1c6119018960011b67ffffffffffffffff1690565b60208b01518b51861867ffffffffffffffff168c5291189190911897508118600181901b603f9190911c18935060c08801518118601481901c602c9190911b1867ffffffffffffffff1660208901526101208801518718602c81901c60149190911b1867ffffffffffffffff1660c08901526102c08801518618600381901c603d9190911b1867ffffffffffffffff166101208901526101c08801518718601981901c60279190911b1867ffffffffffffffff166102c08901526102808801518218602e81901c60129190911b1867ffffffffffffffff166101c089015260408801518618600281901c603e9190911b1867ffffffffffffffff166102808901526101808801518618601581901c602b9190911b1867ffffffffffffffff1660408901526101a08801518518602781901c60199190911b1867ffffffffffffffff166101808901526102608801518718603881901c60089190911b1867ffffffffffffffff166101a08901526102e08801518518600881901c60389190911b1867ffffffffffffffff166102608901526101e08801518218601781901c60299190911b1867ffffffffffffffff166102e089015260808801518718602581901c601b9190911b1867ffffffffffffffff166101e08901526103008801518718603281901c600e9190911b1867ffffffffffffffff1660808901526102a08801518118603e81901c60029190911b1867ffffffffffffffff166103008901526101008801518518600981901c60379190911b1867ffffffffffffffff166102a08901526102008801518118601381901c602d9190911b1867ffffffffffffffff1661010089015260a08801518218601c81901c60249190911b1867ffffffffffffffff1661020089015260608801518518602481901c601c9190911b1867ffffffffffffffff1660a08901526102408801518518602b81901c60159190911b1867ffffffffffffffff1660608901526102208801518618603181901c600f9190911b1867ffffffffffffffff166102408901526101608801518118603681901c600a9190911b1867ffffffffffffffff166102208901525060e08701518518603a81901c60069190911b1867ffffffffffffffff166101608801526101408701518118603d81901c60039190911b1867ffffffffffffffff1660e0880152505067ffffffffffffffff81166101408601525050505050565b611c81816117d4565b805160208201805160408401805160608601805160808801805167ffffffffffffffff871986168a188116808c528619851689188216909952831982169095188516909552841988169091188316909152941990921618811690925260a08301805160c0808601805160e0880180516101008a0180516101208c018051861985168a188d16909a528319821686188c16909652801989169092188a169092528619861618881690529219909216909218841690526101408401805161016086018051610180880180516101a08a0180516101c08c0180518619851689188d169099528319821686188c16909652801988169092188a169092528519851618881690529119909116909118841690526101e08401805161020086018051610220880180516102408a0180516102608c0180518619851689188d169099528319821686188c16909652801988169092188a16909252851985161888169052911990911690911884169052610280840180516102a0860180516102c0880180516102e08a0180516103008c0180518619851689188d169099528319821686188c16909652801988169092188a16909252851985161888169052911990911690911884169052600386901b850151901c908118909116825261160c565b611e6660008284611c78565b611e7260018284611c78565b611e7e60028284611c78565b611e8a60038284611c78565b611e9660048284611c78565b611ea260058284611c78565b611eae60068284611c78565b611eba60078284611c78565b611ec660088284611c78565b611ed260098284611c78565b611ede600a8284611c78565b611eea600b8284611c78565b611ef6600c8284611c78565b611f02600d8284611c78565b611f0e600e8284611c78565b611f1a600f8284611c78565b611f2660108284611c78565b611f3260118284611c78565b611f3e60128284611c78565b611f4a60138284611c78565b611f5660148284611c78565b611f6260158284611c78565b611f6e60168284611c78565b61160c60178284611c78565b7f01000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff831617612020818360408051600093845233602052918152606090922091527effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f01000000000000000000000000000000000000000000000000000000000000001790565b9392505050565b6060604051905081602082018181018286833760888306808015612085576088829003850160808582017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01536001845160001a1784538652612097565b60018353608060878401536088850186525b5050505050601f19603f82510116810160405292915050565b60801c63ffffffff1690565b600061213f565b66ff00ff00ff00ff8160081c1667ff00ff00ff00ff006120ed8360081b67ffffffffffffffff1690565b1617905065ffff0000ffff8160101c1667ffff0000ffff000061211a8360101b67ffffffffffffffff1690565b1617905060008160201c6121388360201b67ffffffffffffffff1690565b1792915050565b60808201516020830190612157906120c3565b6120c3565b6040820151612165906120c3565b60401b1761217d61215260018460059190911b015190565b825160809190911b9061218f906120c3565b60c01b17179392505050565b82601081019282156121c9579160200282015b828111156121c95782518255916020019190600101906121ae565b506121d59291506121f1565b5090565b60405180602001604052806121ec612206565b905290565b5b808211156121d557600081556001016121f2565b6040518061032001604052806019906020820280368337509192915050565b803573ffffffffffffffffffffffffffffffffffffffff8116811461224957600080fd5b919050565b6000806040838503121561226157600080fd5b61226a83612225565b946020939093013593505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b604051610320810167ffffffffffffffff811182821017156122cb576122cb612278565b60405290565b6040516060810167ffffffffffffffff811182821017156122cb576122cb612278565b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016810167ffffffffffffffff8111828210171561233b5761233b612278565b604052919050565b600061032080838503121561235757600080fd5b604051602080820167ffffffffffffffff838210818311171561237c5761237c612278565b8160405283955087601f88011261239257600080fd5b61239a6122a7565b94870194915081888611156123ae57600080fd5b875b868110156123d657803583811681146123c95760008081fd5b84529284019284016123b0565b50909352509295945050505050565b6000606082840312156123f757600080fd5b50919050565b60008083601f84011261240f57600080fd5b50813567ffffffffffffffff81111561242757600080fd5b6020830191508360208260051b850101111561244257600080fd5b9250929050565b60008060008060008060008060006103e08a8c03121561246857600080fd5b6124718a612225565b985060208a013597506124878b60408c01612343565b96506103608a013567ffffffffffffffff808211156124a557600080fd5b6124b18d838e016123e5565b97506103808c01359150808211156124c857600080fd5b6124d48d838e016123fd565b90975095506103a08c01359150808211156124ee57600080fd5b6124fa8d838e016123e5565b94506103c08c013591508082111561251157600080fd5b5061251e8c828d016123fd565b915080935050809150509295985092959850929598565b600080600080600060a0868803121561254d57600080fd5b505083359560208501359550604085013594606081013594506080013592509050565b6000806040838503121561258357600080fd5b50508035926020909101359150565b6000602082840312156125a457600080fd5b5035919050565b60008083601f8401126125bd57600080fd5b50813567ffffffffffffffff8111156125d557600080fd5b60208301915083602082850101111561244257600080fd5b6000806000806000806080878903121561260657600080fd5b86359550602087013567ffffffffffffffff8082111561262557600080fd5b6126318a838b016125ab565b9097509550604089013591508082111561264a57600080fd5b5061265789828a016123fd565b9094509250506060870135801515811461267057600080fd5b809150509295509295509295565b60008060006060848603121561269357600080fd5b61269c84612225565b95602085013595506040909401359392505050565b6000806000604084860312156126c657600080fd5b83359250602084013567ffffffffffffffff8111156126e457600080fd5b6126f0868287016125ab565b9497909650939450505050565b60008060008060006080868803121561271557600080fd5b61271e86612225565b945060208601359350604086013567ffffffffffffffff8082111561274257600080fd5b61274e89838a016123e5565b9450606088013591508082111561276457600080fd5b50612771888289016123fd565b969995985093965092949392505050565b803563ffffffff8116811461224957600080fd5b6000806000606084860312156127ab57600080fd5b833592506127bb60208501612782565b91506127c960408501612782565b90509250925092565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff820361286157612861612801565b5060010190565b60006060823603121561287a57600080fd5b6128826122d1565b823567ffffffffffffffff8082111561289a57600080fd5b9084019036601f8301126128ad57600080fd5b81356020828211156128c1576128c1612278565b6128f1817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f850116016122f4565b9250818352368183860101111561290757600080fd5b81818501828501376000918301810191909152908352848101359083015250604092830135928101929092525090565b81516103208201908260005b601981101561296c57825167ffffffffffffffff16825260209283019290910190600101612943565b50505092915050565b6000821982111561298857612988612801565b500190565b60008083357fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe18436030181126129c257600080fd5b83018035915067ffffffffffffffff8211156129dd57600080fd5b60200191503681900382131561244257600080fd5b600181815b80851115612a4b57817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04821115612a3157612a31612801565b80851615612a3e57918102915b93841c93908002906129f7565b509250929050565b600082612a6257506001612b0e565b81612a6f57506000612b0e565b8160018114612a855760028114612a8f57612aab565b6001915050612b0e565b60ff841115612aa057612aa0612801565b50506001821b612b0e565b5060208310610133831016604e8410600b8410161715612ace575081810a612b0e565b612ad883836129f2565b807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04821115612b0a57612b0a612801565b0290505b92915050565b60006120208383612a53565b600082821015612b3257612b32612801565b500390565b600063ffffffff83811690831681811015612b5457612b54612801565b039392505050565b600063ffffffff808316818516808303821115612b7b57612b7b612801565b01949350505050565b6000845160005b81811015612ba55760208188018101518583015201612b8b565b81811115612bb4576000828501525b509190910192835250602082015260400191905056fea164736f6c634300080f000a"


func init() {
	if err := json.Unmarshal([]byte(PreimageOracleStorageLayoutJSON), PreimageOracleStorageLayout); err != nil {
		panic(err)
	}

	layouts["PreimageOracle"] = PreimageOracleStorageLayout
	deployedBytecodes["PreimageOracle"] = PreimageOracleDeployedBin
	immutableReferences["PreimageOracle"] = false
}
