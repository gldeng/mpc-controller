// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package MpcCoordinator

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// MpcCoordinatorKeyInfo is an auto generated low-level Go binding around an user-defined struct.
type MpcCoordinatorKeyInfo struct {
	GroupId   [32]byte
	Confirmed bool
}

// MpcCoordinatorMetaData contains all meta data concerning the MpcCoordinator contract.
var MpcCoordinatorMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"}],\"name\":\"KeyGenerated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"KeygenRequestAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"ParticipantAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"SignRequestAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"SignRequestStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"StakeRequestAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"participantIndices\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"StakeRequestStarted\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"_calculateAddressForTempTest\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_generatedKeyOnlyForTempTest\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"publicKeys\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"createGroup\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"getGroup\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"participants\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"}],\"name\":\"getKey\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"confirmed\",\"type\":\"bool\"}],\"internalType\":\"structMpcCoordinator.KeyInfo\",\"name\":\"keyInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStakeAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStakeAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStakeNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"myIndex\",\"type\":\"uint256\"}],\"name\":\"joinRequest\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"myIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"generatedPublicKey\",\"type\":\"bytes\"}],\"name\":\"reportGeneratedKey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"requestKeygen\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"requestSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"requestStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"serveStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakeAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakeNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Sigs: map[string]string{
		"303fdc24": "_calculateAddressForTempTest()",
		"ee48981e": "_generatedKeyOnlyForTempTest()",
		"dd6bd149": "createGroup(bytes[],uint256)",
		"12065fe0": "getBalance()",
		"b567d4ba": "getGroup(bytes32)",
		"7fed84f2": "getKey(bytes)",
		"5c98513f": "getStakeAddress()",
		"722580b6": "getStakeAmount()",
		"ff30067e": "getStakeNumber()",
		"2ed92550": "joinRequest(uint256,uint256)",
		"fae3a93c": "reportGeneratedKey(bytes32,uint256,bytes)",
		"e661d90d": "requestKeygen(bytes32)",
		"2f7e3d17": "requestSign(bytes,bytes)",
		"2dbf0344": "requestStake(bytes,string,uint256,uint256,uint256)",
		"6cfb1929": "serveStake(string,uint256,uint256,uint256)",
		"60c7dc47": "stakeAmount()",
		"a85e3863": "stakeNumber()",
	},
	Bin: "0x6080604052611c9c806100136000396000f3fe6080604052600436106101025760003560e01c8063722580b611610095578063dd6bd14911610064578063dd6bd149146102b4578063e661d90d146102d4578063ee48981e146102f4578063fae3a93c14610316578063ff30067e1461033657600080fd5b8063722580b61461021e5780637fed84f214610233578063a85e386314610270578063b567d4ba1461028657600080fd5b8063303fdc24116100d1578063303fdc24146101925780635c98513f146101ca57806360c7dc47146101e85780636cfb1929146101fe57600080fd5b806312065fe01461010e5780632dbf0344146101305780632ed92550146101525780632f7e3d171461017257600080fd5b3661010957005b600080fd5b34801561011a57600080fd5b50475b6040519081526020015b60405180910390f35b34801561013c57600080fd5b5061015061014b366004611500565b61034b565b005b34801561015e57600080fd5b5061015061016d3660046115f5565b610473565b34801561017e57600080fd5b5061015061018d366004611617565b610858565b34801561019e57600080fd5b506001546101b2906001600160a01b031681565b6040516001600160a01b039091168152602001610127565b3480156101d657600080fd5b506001546001600160a01b03166101b2565b3480156101f457600080fd5b5061011d60035481565b34801561020a57600080fd5b50610150610219366004611683565b610948565b34801561022a57600080fd5b5060035461011d565b34801561023f57600080fd5b5061025361024e3660046116dd565b610a5a565b604080518251815260209283015115159281019290925201610127565b34801561027c57600080fd5b5061011d60025481565b34801561029257600080fd5b506102a66102a136600461171f565b610ab2565b604051610127929190611790565b3480156102c057600080fd5b506101506102cf3660046117f9565b610c61565b3480156102e057600080fd5b506101506102ef36600461171f565b610f16565b34801561030057600080fd5b50610309610f44565b6040516101279190611874565b34801561032257600080fd5b50610150610331366004611887565b610fd2565b34801561034257600080fd5b5060025461011d565b600060078760405161035d91906118ce565b90815260408051602092819003830181208183019092528154815260019091015460ff16151591810182905291506103b05760405162461bcd60e51b81526004016103a7906118ea565b60405180910390fd5b60006103ba611179565b60008181526009602090815260409091208a51929350916103e0918391908c0190611394565b506000828152600a602052604090206103fa818a8a611418565b5060018101879055600281018690556003810185905560405161041e908b906118ce565b60405180910390207f18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594848b8b8b8b8b60405161045f9695949392919061195f565b60405180910390a250505050505050505050565b60008281526009602052604081208054909190829061049190611997565b9050116104d95760405162461bcd60e51b81526020600482015260166024820152752932b8bab2b9ba103237b2b9b713ba1032bc34b9ba1760511b60448201526064016103a7565b60006007826000016040516104ee91906119d1565b90815260408051602092819003830181208183019092528154815260019091015460ff16151591810182905291506105845760405162461bcd60e51b815260206004820152603360248201527f5075626c6963206b657920646f65736e2774206578697374206f7220686173206044820152723737ba103132b2b71031b7b73334b936b2b21760691b60648201526084016103a7565b805160009081526005602052604090205460028301548110156105e05760405162461bcd60e51b815260206004820152601460248201527321b0b73737ba103537b4b71030b73cb6b7b9329760611b60448201526064016103a7565b81516105ec908561119b565b60005b600284015481101561066d578484600201828154811061061157610611611a43565b90600052602060002001540361065b5760405162461bcd60e51b815260206004820152600f60248201526e20b63932b0b23c903537b4b732b21760891b60448201526064016103a7565b8061066581611a6f565b9150506105ef565b50600283018054600181810183556000928352602090922001859055610694908290611a88565b6002840154036108515760008360010180546106af90611997565b9050111561070e576040516106c59084906119d1565b60405180910390207f279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad3038685600101604051610701929190611aa0565b60405180910390a2610851565b6000858152600a602052604080822081516080810190925280548290829061073590611997565b80601f016020809104026020016040519081016040528092919081815260200182805461076190611997565b80156107ae5780601f10610783576101008083540402835291602001916107ae565b820191906000526020600020905b81548152906001019060200180831161079157829003601f168201915b505050505081526020016001820154815260200160028201548152602001600382015481525050905060008160200151111561084f576040516107f29085906119d1565b60405180910390207f288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f304868786600201846000015185602001518660400151876060015160405161084696959493929190611b2d565b60405180910390a25b505b5050505050565b60006007858560405161086c929190611ba8565b90815260408051602092819003830181208183019092528154815260019091015460ff16151591810182905291506108b65760405162461bcd60e51b81526004016103a7906118ea565b60006108c0611179565b60008181526009602052604090209091506108dc818888611418565b506108eb600182018686611418565b5086866040516108fc929190611ba8565b60405180910390207ffd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b2262883878760405161093793929190611bb8565b60405180910390a250505050505050565b600080805461095690611997565b80601f016020809104026020016040519081016040528092919081815260200182805461098290611997565b80156109cf5780601f106109a4576101008083540402835291602001916109cf565b820191906000526020600020905b8154815290600101906020018083116109b257829003601f168201915b50506001546040519495506001600160a01b03169384935088156108fc0292508891506000818181858888f19350505050158015610a11573d6000803e3d6000fd5b50610a2082888888888861034b565b600160026000828254610a339190611a88565b925050819055508460036000828254610a4c9190611a88565b909155505050505050505050565b604080518082019091526000808252602082015260078383604051610a80929190611ba8565b9081526040805191829003602090810183208383019092528154835260019091015460ff161515908201529392505050565b6000818152600460205260408120546060919080610b095760405162461bcd60e51b815260206004820152601460248201527323b937bab8103237b2b9b713ba1032bc34b9ba1760611b60448201526064016103a7565b60008167ffffffffffffffff811115610b2457610b246114a1565b604051908082528060200260200182016040528015610b5757816020015b6060815260200190600190039081610b425790505b5060008681526005602052604081205494509091505b82811015610c5657600086815260066020526040812090610b8f836001611a88565b81526020019081526020016000208054610ba890611997565b80601f0160208091040260200160405190810160405280929190818152602001828054610bd490611997565b8015610c215780601f10610bf657610100808354040283529160200191610c21565b820191906000526020600020905b815481529060010190602001808311610c0457829003601f168201915b5050505050828281518110610c3857610c38611a43565b60200260200101819052508080610c4e90611a6f565b915050610b6d565b508093505050915091565b60018211610cc25760405162461bcd60e51b815260206004820152602860248201527f412067726f75702072657175697265732032206f72206d6f726520706172746960448201526731b4b830b73a399760c11b60648201526084016103a7565b60018110158015610cd257508181105b610d125760405162461bcd60e51b8152602060048201526011602482015270125b9d985b1a59081d1a1c995cda1bdb19607a1b60448201526064016103a7565b604080516020810183905260009101604051602081830303815290604052905060005b83811015610d985781858583818110610d5057610d50611a43565b9050602002810190610d629190611bdb565b604051602001610d7493929190611c22565b60405160208183030381529060405291508080610d9090611a6f565b915050610d35565b508051602080830191909120600081815260049092526040909120548015610dfa5760405162461bcd60e51b815260206004820152601560248201527423b937bab81030b63932b0b23c9032bc34b9ba399760591b60448201526064016103a7565b6000828152600460209081526040808320889055600590915281208590555b85811015610f0d57868682818110610e3357610e33611a43565b9050602002810190610e459190611bdb565b600085815260066020526040812090610e5f856001611a88565b81526020019081526020016000209190610e7a929190611418565b50868682818110610e8d57610e8d611a43565b9050602002810190610e9f9190611bdb565b604051610ead929190611ba8565b6040519081900390207f39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a84610ee3846001611a88565b6040805192835260208301919091520160405180910390a280610f0581611a6f565b915050610e19565b50505050505050565b60405181907f5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa732790600090a250565b60008054610f5190611997565b80601f0160208091040260200160405190810160405280929190818152602001828054610f7d90611997565b8015610fca5780601f10610f9f57610100808354040283529160200191610fca565b820191906000526020600020905b815481529060010190602001808311610fad57829003601f168201915b505050505081565b8383610fde828261119b565b600060078585604051610ff2929190611ba8565b908152604051908190036020019020600181015490915060ff16156110755760405162461bcd60e51b815260206004820152603360248201527f4b65792068617320616c7265616479206265656e20636f6e6669726d656420626044820152723c9030b636103830b93a34b1b4b830b73a399760691b60648201526084016103a7565b600160088686604051611089929190611ba8565b908152604080516020928190038301902060008a815292529020805460ff19169115159190911790556110bd8786866112fa565b15610f0d578681556001808201805460ff191690911790556110e160008686611418565b5061112185858080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061138492505050565b600160006101000a8154816001600160a01b0302191690836001600160a01b03160217905550867f767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d8686604051610937929190611c4a565b60006001600b600082825461118e9190611a88565b9091555050600b54919050565b6000828152600660209081526040808320848452909152812080546111bf90611997565b80601f01602080910402602001604051908101604052809291908181526020018280546111eb90611997565b80156112385780601f1061120d57610100808354040283529160200191611238565b820191906000526020600020905b81548152906001019060200180831161121b57829003601f168201915b5050505050905060008151116112905760405162461bcd60e51b815260206004820152601960248201527f496e76616c69642067726f75704964206f7220696e6465782e0000000000000060448201526064016103a7565b8051602082012060008190526001600160a01b03811633146112f45760405162461bcd60e51b815260206004820152601c60248201527f43616c6c6572206973206e6f7420612067726f7570206d656d6265720000000060448201526064016103a7565b50505050565b600083815260046020526040812054815b818110156113765760088585604051611325929190611ba8565b9081526040519081900360200190206000611341836001611a88565b815260208101919091526040016000205460ff166113645760009250505061137d565b8061136e81611a6f565b91505061130b565b5060019150505b9392505050565b8051602090910120600081905290565b8280546113a090611997565b90600052602060002090601f0160209004810192826113c25760008555611408565b82601f106113db57805160ff1916838001178555611408565b82800160010185558215611408579182015b828111156114085782518255916020019190600101906113ed565b5061141492915061148c565b5090565b82805461142490611997565b90600052602060002090601f0160209004810192826114465760008555611408565b82601f1061145f5782800160ff19823516178555611408565b82800160010185558215611408579182015b82811115611408578235825591602001919060010190611471565b5b80821115611414576000815560010161148d565b634e487b7160e01b600052604160045260246000fd5b60008083601f8401126114c957600080fd5b50813567ffffffffffffffff8111156114e157600080fd5b6020830191508360208285010111156114f957600080fd5b9250929050565b60008060008060008060a0878903121561151957600080fd5b863567ffffffffffffffff8082111561153157600080fd5b818901915089601f83011261154557600080fd5b813581811115611557576115576114a1565b604051601f8201601f19908116603f0116810190838211818310171561157f5761157f6114a1565b816040528281528c602084870101111561159857600080fd5b82602086016020830137600060208483010152809a5050505060208901359150808211156115c557600080fd5b506115d289828a016114b7565b979a90995096976040810135976060820135975060809091013595509350505050565b6000806040838503121561160857600080fd5b50508035926020909101359150565b6000806000806040858703121561162d57600080fd5b843567ffffffffffffffff8082111561164557600080fd5b611651888389016114b7565b9096509450602087013591508082111561166a57600080fd5b50611677878288016114b7565b95989497509550505050565b60008060008060006080868803121561169b57600080fd5b853567ffffffffffffffff8111156116b257600080fd5b6116be888289016114b7565b9099909850602088013597604081013597506060013595509350505050565b600080602083850312156116f057600080fd5b823567ffffffffffffffff81111561170757600080fd5b611713858286016114b7565b90969095509350505050565b60006020828403121561173157600080fd5b5035919050565b60005b8381101561175357818101518382015260200161173b565b838111156112f45750506000910152565b6000815180845261177c816020860160208601611738565b601f01601f19169290920160200192915050565b6000604082016040835280855180835260608501915060608160051b8601019250602080880160005b838110156117e757605f198887030185526117d5868351611764565b955093820193908201906001016117b9565b50509490940194909452949350505050565b60008060006040848603121561180e57600080fd5b833567ffffffffffffffff8082111561182657600080fd5b818601915086601f83011261183a57600080fd5b81358181111561184957600080fd5b8760208260051b850101111561185e57600080fd5b6020928301989097509590910135949350505050565b60208152600061137d6020830184611764565b6000806000806060858703121561189d57600080fd5b8435935060208501359250604085013567ffffffffffffffff8111156118c257600080fd5b611677878288016114b7565b600082516118e0818460208701611738565b9190910192915050565b6020808252602c908201527f4b657920646f65736e2774206578697374206f7220686173206e6f742062656560408201526b371031b7b73334b936b2b21760a11b606082015260800190565b81835281816020850137506000828201602090810191909152601f909101601f19169091010190565b86815260a06020820152600061197960a083018789611936565b60408301959095525060608101929092526080909101529392505050565b600181811c908216806119ab57607f821691505b6020821081036119cb57634e487b7160e01b600052602260045260246000fd5b50919050565b60008083546119df81611997565b600182811680156119f75760018114611a0857611a37565b60ff19841687528287019450611a37565b8760005260208060002060005b85811015611a2e5781548a820152908401908201611a15565b50505082870194505b50929695505050505050565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052601160045260246000fd5b600060018201611a8157611a81611a59565b5060010190565b60008219821115611a9b57611a9b611a59565b500190565b8281526000602060408184015260008454611aba81611997565b8060408701526060600180841660008114611adc5760018114611af057611b1e565b60ff19851689840152608089019550611b1e565b896000528660002060005b85811015611b165781548b8201860152908301908801611afb565b8a0184019650505b50939998505050505050505050565b600060c08201888352602060c08185015281895480845260e0860191508a60005282600020935060005b81811015611b7357845483526001948501949284019201611b57565b50508481036040860152611b87818a611764565b606086019890985250505050608081019290925260a0909101529392505050565b8183823760009101908152919050565b838152604060208201526000611bd2604083018486611936565b95945050505050565b6000808335601e19843603018112611bf257600080fd5b83018035915067ffffffffffffffff821115611c0d57600080fd5b6020019150368190038213156114f957600080fd5b60008451611c34818460208901611738565b8201838582376000930192835250909392505050565b602081526000611c5e602083018486611936565b94935050505056fea26469706673582212207596a6155c0068db1a1acc1e81fe24c74ee71e0b6d365a611020591f0e27854e64736f6c634300080d0033",
}

// MpcCoordinatorABI is the input ABI used to generate the binding from.
// Deprecated: Use MpcCoordinatorMetaData.ABI instead.
var MpcCoordinatorABI = MpcCoordinatorMetaData.ABI

// Deprecated: Use MpcCoordinatorMetaData.Sigs instead.
// MpcCoordinatorFuncSigs maps the 4-byte function signature to its string representation.
var MpcCoordinatorFuncSigs = MpcCoordinatorMetaData.Sigs

// MpcCoordinatorBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use MpcCoordinatorMetaData.Bin instead.
var MpcCoordinatorBin = MpcCoordinatorMetaData.Bin

// DeployMpcCoordinator deploys a new Ethereum contract, binding an instance of MpcCoordinator to it.
func DeployMpcCoordinator(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *MpcCoordinator, error) {
	parsed, err := MpcCoordinatorMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(MpcCoordinatorBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MpcCoordinator{MpcCoordinatorCaller: MpcCoordinatorCaller{contract: contract}, MpcCoordinatorTransactor: MpcCoordinatorTransactor{contract: contract}, MpcCoordinatorFilterer: MpcCoordinatorFilterer{contract: contract}}, nil
}

// MpcCoordinator is an auto generated Go binding around an Ethereum contract.
type MpcCoordinator struct {
	MpcCoordinatorCaller     // Read-only binding to the contract
	MpcCoordinatorTransactor // Write-only binding to the contract
	MpcCoordinatorFilterer   // Log filterer for contract events
}

// MpcCoordinatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type MpcCoordinatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MpcCoordinatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MpcCoordinatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MpcCoordinatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MpcCoordinatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MpcCoordinatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MpcCoordinatorSession struct {
	Contract     *MpcCoordinator   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MpcCoordinatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MpcCoordinatorCallerSession struct {
	Contract *MpcCoordinatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// MpcCoordinatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MpcCoordinatorTransactorSession struct {
	Contract     *MpcCoordinatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// MpcCoordinatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type MpcCoordinatorRaw struct {
	Contract *MpcCoordinator // Generic contract binding to access the raw methods on
}

// MpcCoordinatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MpcCoordinatorCallerRaw struct {
	Contract *MpcCoordinatorCaller // Generic read-only contract binding to access the raw methods on
}

// MpcCoordinatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MpcCoordinatorTransactorRaw struct {
	Contract *MpcCoordinatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMpcCoordinator creates a new instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinator(address common.Address, backend bind.ContractBackend) (*MpcCoordinator, error) {
	contract, err := bindMpcCoordinator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinator{MpcCoordinatorCaller: MpcCoordinatorCaller{contract: contract}, MpcCoordinatorTransactor: MpcCoordinatorTransactor{contract: contract}, MpcCoordinatorFilterer: MpcCoordinatorFilterer{contract: contract}}, nil
}

// NewMpcCoordinatorCaller creates a new read-only instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinatorCaller(address common.Address, caller bind.ContractCaller) (*MpcCoordinatorCaller, error) {
	contract, err := bindMpcCoordinator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorCaller{contract: contract}, nil
}

// NewMpcCoordinatorTransactor creates a new write-only instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinatorTransactor(address common.Address, transactor bind.ContractTransactor) (*MpcCoordinatorTransactor, error) {
	contract, err := bindMpcCoordinator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorTransactor{contract: contract}, nil
}

// NewMpcCoordinatorFilterer creates a new log filterer instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinatorFilterer(address common.Address, filterer bind.ContractFilterer) (*MpcCoordinatorFilterer, error) {
	contract, err := bindMpcCoordinator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorFilterer{contract: contract}, nil
}

// bindMpcCoordinator binds a generic wrapper to an already deployed contract.
func bindMpcCoordinator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MpcCoordinatorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MpcCoordinator *MpcCoordinatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MpcCoordinator.Contract.MpcCoordinatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MpcCoordinator *MpcCoordinatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.MpcCoordinatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MpcCoordinator *MpcCoordinatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.MpcCoordinatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MpcCoordinator *MpcCoordinatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MpcCoordinator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MpcCoordinator *MpcCoordinatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MpcCoordinator *MpcCoordinatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.contract.Transact(opts, method, params...)
}

// CalculateAddressForTempTest is a free data retrieval call binding the contract method 0x303fdc24.
//
// Solidity: function _calculateAddressForTempTest() view returns(address)
func (_MpcCoordinator *MpcCoordinatorCaller) CalculateAddressForTempTest(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "_calculateAddressForTempTest")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CalculateAddressForTempTest is a free data retrieval call binding the contract method 0x303fdc24.
//
// Solidity: function _calculateAddressForTempTest() view returns(address)
func (_MpcCoordinator *MpcCoordinatorSession) CalculateAddressForTempTest() (common.Address, error) {
	return _MpcCoordinator.Contract.CalculateAddressForTempTest(&_MpcCoordinator.CallOpts)
}

// CalculateAddressForTempTest is a free data retrieval call binding the contract method 0x303fdc24.
//
// Solidity: function _calculateAddressForTempTest() view returns(address)
func (_MpcCoordinator *MpcCoordinatorCallerSession) CalculateAddressForTempTest() (common.Address, error) {
	return _MpcCoordinator.Contract.CalculateAddressForTempTest(&_MpcCoordinator.CallOpts)
}

// GeneratedKeyOnlyForTempTest is a free data retrieval call binding the contract method 0xee48981e.
//
// Solidity: function _generatedKeyOnlyForTempTest() view returns(bytes)
func (_MpcCoordinator *MpcCoordinatorCaller) GeneratedKeyOnlyForTempTest(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "_generatedKeyOnlyForTempTest")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GeneratedKeyOnlyForTempTest is a free data retrieval call binding the contract method 0xee48981e.
//
// Solidity: function _generatedKeyOnlyForTempTest() view returns(bytes)
func (_MpcCoordinator *MpcCoordinatorSession) GeneratedKeyOnlyForTempTest() ([]byte, error) {
	return _MpcCoordinator.Contract.GeneratedKeyOnlyForTempTest(&_MpcCoordinator.CallOpts)
}

// GeneratedKeyOnlyForTempTest is a free data retrieval call binding the contract method 0xee48981e.
//
// Solidity: function _generatedKeyOnlyForTempTest() view returns(bytes)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GeneratedKeyOnlyForTempTest() ([]byte, error) {
	return _MpcCoordinator.Contract.GeneratedKeyOnlyForTempTest(&_MpcCoordinator.CallOpts)
}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCaller) GetBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorSession) GetBalance() (*big.Int, error) {
	return _MpcCoordinator.Contract.GetBalance(&_MpcCoordinator.CallOpts)
}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetBalance() (*big.Int, error) {
	return _MpcCoordinator.Contract.GetBalance(&_MpcCoordinator.CallOpts)
}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_MpcCoordinator *MpcCoordinatorCaller) GetGroup(opts *bind.CallOpts, groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getGroup", groupId)

	outstruct := new(struct {
		Participants [][]byte
		Threshold    *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Participants = *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)
	outstruct.Threshold = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_MpcCoordinator *MpcCoordinatorSession) GetGroup(groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	return _MpcCoordinator.Contract.GetGroup(&_MpcCoordinator.CallOpts, groupId)
}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetGroup(groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	return _MpcCoordinator.Contract.GetGroup(&_MpcCoordinator.CallOpts, groupId)
}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_MpcCoordinator *MpcCoordinatorCaller) GetKey(opts *bind.CallOpts, publicKey []byte) (MpcCoordinatorKeyInfo, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getKey", publicKey)

	if err != nil {
		return *new(MpcCoordinatorKeyInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(MpcCoordinatorKeyInfo)).(*MpcCoordinatorKeyInfo)

	return out0, err

}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_MpcCoordinator *MpcCoordinatorSession) GetKey(publicKey []byte) (MpcCoordinatorKeyInfo, error) {
	return _MpcCoordinator.Contract.GetKey(&_MpcCoordinator.CallOpts, publicKey)
}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetKey(publicKey []byte) (MpcCoordinatorKeyInfo, error) {
	return _MpcCoordinator.Contract.GetKey(&_MpcCoordinator.CallOpts, publicKey)
}

// GetStakeAddress is a free data retrieval call binding the contract method 0x5c98513f.
//
// Solidity: function getStakeAddress() view returns(address)
func (_MpcCoordinator *MpcCoordinatorCaller) GetStakeAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getStakeAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetStakeAddress is a free data retrieval call binding the contract method 0x5c98513f.
//
// Solidity: function getStakeAddress() view returns(address)
func (_MpcCoordinator *MpcCoordinatorSession) GetStakeAddress() (common.Address, error) {
	return _MpcCoordinator.Contract.GetStakeAddress(&_MpcCoordinator.CallOpts)
}

// GetStakeAddress is a free data retrieval call binding the contract method 0x5c98513f.
//
// Solidity: function getStakeAddress() view returns(address)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetStakeAddress() (common.Address, error) {
	return _MpcCoordinator.Contract.GetStakeAddress(&_MpcCoordinator.CallOpts)
}

// GetStakeAmount is a free data retrieval call binding the contract method 0x722580b6.
//
// Solidity: function getStakeAmount() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCaller) GetStakeAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getStakeAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetStakeAmount is a free data retrieval call binding the contract method 0x722580b6.
//
// Solidity: function getStakeAmount() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorSession) GetStakeAmount() (*big.Int, error) {
	return _MpcCoordinator.Contract.GetStakeAmount(&_MpcCoordinator.CallOpts)
}

// GetStakeAmount is a free data retrieval call binding the contract method 0x722580b6.
//
// Solidity: function getStakeAmount() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetStakeAmount() (*big.Int, error) {
	return _MpcCoordinator.Contract.GetStakeAmount(&_MpcCoordinator.CallOpts)
}

// GetStakeNumber is a free data retrieval call binding the contract method 0xff30067e.
//
// Solidity: function getStakeNumber() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCaller) GetStakeNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getStakeNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetStakeNumber is a free data retrieval call binding the contract method 0xff30067e.
//
// Solidity: function getStakeNumber() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorSession) GetStakeNumber() (*big.Int, error) {
	return _MpcCoordinator.Contract.GetStakeNumber(&_MpcCoordinator.CallOpts)
}

// GetStakeNumber is a free data retrieval call binding the contract method 0xff30067e.
//
// Solidity: function getStakeNumber() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetStakeNumber() (*big.Int, error) {
	return _MpcCoordinator.Contract.GetStakeNumber(&_MpcCoordinator.CallOpts)
}

// StakeAmount is a free data retrieval call binding the contract method 0x60c7dc47.
//
// Solidity: function stakeAmount() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCaller) StakeAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "stakeAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeAmount is a free data retrieval call binding the contract method 0x60c7dc47.
//
// Solidity: function stakeAmount() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorSession) StakeAmount() (*big.Int, error) {
	return _MpcCoordinator.Contract.StakeAmount(&_MpcCoordinator.CallOpts)
}

// StakeAmount is a free data retrieval call binding the contract method 0x60c7dc47.
//
// Solidity: function stakeAmount() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCallerSession) StakeAmount() (*big.Int, error) {
	return _MpcCoordinator.Contract.StakeAmount(&_MpcCoordinator.CallOpts)
}

// StakeNumber is a free data retrieval call binding the contract method 0xa85e3863.
//
// Solidity: function stakeNumber() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCaller) StakeNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "stakeNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeNumber is a free data retrieval call binding the contract method 0xa85e3863.
//
// Solidity: function stakeNumber() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorSession) StakeNumber() (*big.Int, error) {
	return _MpcCoordinator.Contract.StakeNumber(&_MpcCoordinator.CallOpts)
}

// StakeNumber is a free data retrieval call binding the contract method 0xa85e3863.
//
// Solidity: function stakeNumber() view returns(uint256)
func (_MpcCoordinator *MpcCoordinatorCallerSession) StakeNumber() (*big.Int, error) {
	return _MpcCoordinator.Contract.StakeNumber(&_MpcCoordinator.CallOpts)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) CreateGroup(opts *bind.TransactOpts, publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "createGroup", publicKeys, threshold)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_MpcCoordinator *MpcCoordinatorSession) CreateGroup(publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.CreateGroup(&_MpcCoordinator.TransactOpts, publicKeys, threshold)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) CreateGroup(publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.CreateGroup(&_MpcCoordinator.TransactOpts, publicKeys, threshold)
}

// JoinRequest is a paid mutator transaction binding the contract method 0x2ed92550.
//
// Solidity: function joinRequest(uint256 requestId, uint256 myIndex) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) JoinRequest(opts *bind.TransactOpts, requestId *big.Int, myIndex *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "joinRequest", requestId, myIndex)
}

// JoinRequest is a paid mutator transaction binding the contract method 0x2ed92550.
//
// Solidity: function joinRequest(uint256 requestId, uint256 myIndex) returns()
func (_MpcCoordinator *MpcCoordinatorSession) JoinRequest(requestId *big.Int, myIndex *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.JoinRequest(&_MpcCoordinator.TransactOpts, requestId, myIndex)
}

// JoinRequest is a paid mutator transaction binding the contract method 0x2ed92550.
//
// Solidity: function joinRequest(uint256 requestId, uint256 myIndex) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) JoinRequest(requestId *big.Int, myIndex *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.JoinRequest(&_MpcCoordinator.TransactOpts, requestId, myIndex)
}

// ReportGeneratedKey is a paid mutator transaction binding the contract method 0xfae3a93c.
//
// Solidity: function reportGeneratedKey(bytes32 groupId, uint256 myIndex, bytes generatedPublicKey) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) ReportGeneratedKey(opts *bind.TransactOpts, groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "reportGeneratedKey", groupId, myIndex, generatedPublicKey)
}

// ReportGeneratedKey is a paid mutator transaction binding the contract method 0xfae3a93c.
//
// Solidity: function reportGeneratedKey(bytes32 groupId, uint256 myIndex, bytes generatedPublicKey) returns()
func (_MpcCoordinator *MpcCoordinatorSession) ReportGeneratedKey(groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.ReportGeneratedKey(&_MpcCoordinator.TransactOpts, groupId, myIndex, generatedPublicKey)
}

// ReportGeneratedKey is a paid mutator transaction binding the contract method 0xfae3a93c.
//
// Solidity: function reportGeneratedKey(bytes32 groupId, uint256 myIndex, bytes generatedPublicKey) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) ReportGeneratedKey(groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.ReportGeneratedKey(&_MpcCoordinator.TransactOpts, groupId, myIndex, generatedPublicKey)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) RequestKeygen(opts *bind.TransactOpts, groupId [32]byte) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "requestKeygen", groupId)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_MpcCoordinator *MpcCoordinatorSession) RequestKeygen(groupId [32]byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestKeygen(&_MpcCoordinator.TransactOpts, groupId)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) RequestKeygen(groupId [32]byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestKeygen(&_MpcCoordinator.TransactOpts, groupId)
}

// RequestSign is a paid mutator transaction binding the contract method 0x2f7e3d17.
//
// Solidity: function requestSign(bytes publicKey, bytes message) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) RequestSign(opts *bind.TransactOpts, publicKey []byte, message []byte) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "requestSign", publicKey, message)
}

// RequestSign is a paid mutator transaction binding the contract method 0x2f7e3d17.
//
// Solidity: function requestSign(bytes publicKey, bytes message) returns()
func (_MpcCoordinator *MpcCoordinatorSession) RequestSign(publicKey []byte, message []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestSign(&_MpcCoordinator.TransactOpts, publicKey, message)
}

// RequestSign is a paid mutator transaction binding the contract method 0x2f7e3d17.
//
// Solidity: function requestSign(bytes publicKey, bytes message) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) RequestSign(publicKey []byte, message []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestSign(&_MpcCoordinator.TransactOpts, publicKey, message)
}

// RequestStake is a paid mutator transaction binding the contract method 0x2dbf0344.
//
// Solidity: function requestStake(bytes publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) RequestStake(opts *bind.TransactOpts, publicKey []byte, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "requestStake", publicKey, nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x2dbf0344.
//
// Solidity: function requestStake(bytes publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorSession) RequestStake(publicKey []byte, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestStake(&_MpcCoordinator.TransactOpts, publicKey, nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x2dbf0344.
//
// Solidity: function requestStake(bytes publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) RequestStake(publicKey []byte, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestStake(&_MpcCoordinator.TransactOpts, publicKey, nodeID, amount, startTime, endTime)
}

// ServeStake is a paid mutator transaction binding the contract method 0x6cfb1929.
//
// Solidity: function serveStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) ServeStake(opts *bind.TransactOpts, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "serveStake", nodeID, amount, startTime, endTime)
}

// ServeStake is a paid mutator transaction binding the contract method 0x6cfb1929.
//
// Solidity: function serveStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorSession) ServeStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.ServeStake(&_MpcCoordinator.TransactOpts, nodeID, amount, startTime, endTime)
}

// ServeStake is a paid mutator transaction binding the contract method 0x6cfb1929.
//
// Solidity: function serveStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) ServeStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.ServeStake(&_MpcCoordinator.TransactOpts, nodeID, amount, startTime, endTime)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MpcCoordinator.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MpcCoordinator *MpcCoordinatorSession) Receive() (*types.Transaction, error) {
	return _MpcCoordinator.Contract.Receive(&_MpcCoordinator.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) Receive() (*types.Transaction, error) {
	return _MpcCoordinator.Contract.Receive(&_MpcCoordinator.TransactOpts)
}

// MpcCoordinatorKeyGeneratedIterator is returned from FilterKeyGenerated and is used to iterate over the raw logs and unpacked data for KeyGenerated events raised by the MpcCoordinator contract.
type MpcCoordinatorKeyGeneratedIterator struct {
	Event *MpcCoordinatorKeyGenerated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorKeyGeneratedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorKeyGenerated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorKeyGenerated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorKeyGeneratedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorKeyGeneratedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorKeyGenerated represents a KeyGenerated event raised by the MpcCoordinator contract.
type MpcCoordinatorKeyGenerated struct {
	GroupId   [32]byte
	PublicKey []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterKeyGenerated is a free log retrieval operation binding the contract event 0x767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d.
//
// Solidity: event KeyGenerated(bytes32 indexed groupId, bytes publicKey)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterKeyGenerated(opts *bind.FilterOpts, groupId [][32]byte) (*MpcCoordinatorKeyGeneratedIterator, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "KeyGenerated", groupIdRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorKeyGeneratedIterator{contract: _MpcCoordinator.contract, event: "KeyGenerated", logs: logs, sub: sub}, nil
}

// WatchKeyGenerated is a free log subscription operation binding the contract event 0x767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d.
//
// Solidity: event KeyGenerated(bytes32 indexed groupId, bytes publicKey)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchKeyGenerated(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorKeyGenerated, groupId [][32]byte) (event.Subscription, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "KeyGenerated", groupIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorKeyGenerated)
				if err := _MpcCoordinator.contract.UnpackLog(event, "KeyGenerated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseKeyGenerated is a log parse operation binding the contract event 0x767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d.
//
// Solidity: event KeyGenerated(bytes32 indexed groupId, bytes publicKey)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseKeyGenerated(log types.Log) (*MpcCoordinatorKeyGenerated, error) {
	event := new(MpcCoordinatorKeyGenerated)
	if err := _MpcCoordinator.contract.UnpackLog(event, "KeyGenerated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorKeygenRequestAddedIterator is returned from FilterKeygenRequestAdded and is used to iterate over the raw logs and unpacked data for KeygenRequestAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorKeygenRequestAddedIterator struct {
	Event *MpcCoordinatorKeygenRequestAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorKeygenRequestAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorKeygenRequestAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorKeygenRequestAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorKeygenRequestAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorKeygenRequestAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorKeygenRequestAdded represents a KeygenRequestAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorKeygenRequestAdded struct {
	GroupId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterKeygenRequestAdded is a free log retrieval operation binding the contract event 0x5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa7327.
//
// Solidity: event KeygenRequestAdded(bytes32 indexed groupId)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterKeygenRequestAdded(opts *bind.FilterOpts, groupId [][32]byte) (*MpcCoordinatorKeygenRequestAddedIterator, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "KeygenRequestAdded", groupIdRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorKeygenRequestAddedIterator{contract: _MpcCoordinator.contract, event: "KeygenRequestAdded", logs: logs, sub: sub}, nil
}

// WatchKeygenRequestAdded is a free log subscription operation binding the contract event 0x5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa7327.
//
// Solidity: event KeygenRequestAdded(bytes32 indexed groupId)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchKeygenRequestAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorKeygenRequestAdded, groupId [][32]byte) (event.Subscription, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "KeygenRequestAdded", groupIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorKeygenRequestAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "KeygenRequestAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseKeygenRequestAdded is a log parse operation binding the contract event 0x5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa7327.
//
// Solidity: event KeygenRequestAdded(bytes32 indexed groupId)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseKeygenRequestAdded(log types.Log) (*MpcCoordinatorKeygenRequestAdded, error) {
	event := new(MpcCoordinatorKeygenRequestAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "KeygenRequestAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorParticipantAddedIterator is returned from FilterParticipantAdded and is used to iterate over the raw logs and unpacked data for ParticipantAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorParticipantAddedIterator struct {
	Event *MpcCoordinatorParticipantAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorParticipantAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorParticipantAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorParticipantAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorParticipantAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorParticipantAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorParticipantAdded represents a ParticipantAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorParticipantAdded struct {
	PublicKey common.Hash
	GroupId   [32]byte
	Index     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterParticipantAdded is a free log retrieval operation binding the contract event 0x39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a.
//
// Solidity: event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterParticipantAdded(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorParticipantAddedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "ParticipantAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorParticipantAddedIterator{contract: _MpcCoordinator.contract, event: "ParticipantAdded", logs: logs, sub: sub}, nil
}

// WatchParticipantAdded is a free log subscription operation binding the contract event 0x39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a.
//
// Solidity: event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchParticipantAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorParticipantAdded, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "ParticipantAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorParticipantAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "ParticipantAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseParticipantAdded is a log parse operation binding the contract event 0x39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a.
//
// Solidity: event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseParticipantAdded(log types.Log) (*MpcCoordinatorParticipantAdded, error) {
	event := new(MpcCoordinatorParticipantAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "ParticipantAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorSignRequestAddedIterator is returned from FilterSignRequestAdded and is used to iterate over the raw logs and unpacked data for SignRequestAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestAddedIterator struct {
	Event *MpcCoordinatorSignRequestAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorSignRequestAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorSignRequestAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorSignRequestAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorSignRequestAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorSignRequestAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorSignRequestAdded represents a SignRequestAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestAdded struct {
	RequestId *big.Int
	PublicKey common.Hash
	Message   []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSignRequestAdded is a free log retrieval operation binding the contract event 0xfd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b22628.
//
// Solidity: event SignRequestAdded(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterSignRequestAdded(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorSignRequestAddedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "SignRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorSignRequestAddedIterator{contract: _MpcCoordinator.contract, event: "SignRequestAdded", logs: logs, sub: sub}, nil
}

// WatchSignRequestAdded is a free log subscription operation binding the contract event 0xfd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b22628.
//
// Solidity: event SignRequestAdded(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchSignRequestAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorSignRequestAdded, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "SignRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorSignRequestAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignRequestAdded is a log parse operation binding the contract event 0xfd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b22628.
//
// Solidity: event SignRequestAdded(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseSignRequestAdded(log types.Log) (*MpcCoordinatorSignRequestAdded, error) {
	event := new(MpcCoordinatorSignRequestAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorSignRequestStartedIterator is returned from FilterSignRequestStarted and is used to iterate over the raw logs and unpacked data for SignRequestStarted events raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestStartedIterator struct {
	Event *MpcCoordinatorSignRequestStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorSignRequestStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorSignRequestStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorSignRequestStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorSignRequestStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorSignRequestStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorSignRequestStarted represents a SignRequestStarted event raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestStarted struct {
	RequestId *big.Int
	PublicKey common.Hash
	Message   []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSignRequestStarted is a free log retrieval operation binding the contract event 0x279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad303.
//
// Solidity: event SignRequestStarted(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterSignRequestStarted(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorSignRequestStartedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "SignRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorSignRequestStartedIterator{contract: _MpcCoordinator.contract, event: "SignRequestStarted", logs: logs, sub: sub}, nil
}

// WatchSignRequestStarted is a free log subscription operation binding the contract event 0x279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad303.
//
// Solidity: event SignRequestStarted(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchSignRequestStarted(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorSignRequestStarted, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "SignRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorSignRequestStarted)
				if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignRequestStarted is a log parse operation binding the contract event 0x279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad303.
//
// Solidity: event SignRequestStarted(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseSignRequestStarted(log types.Log) (*MpcCoordinatorSignRequestStarted, error) {
	event := new(MpcCoordinatorSignRequestStarted)
	if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorStakeRequestAddedIterator is returned from FilterStakeRequestAdded and is used to iterate over the raw logs and unpacked data for StakeRequestAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestAddedIterator struct {
	Event *MpcCoordinatorStakeRequestAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorStakeRequestAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorStakeRequestAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorStakeRequestAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorStakeRequestAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorStakeRequestAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorStakeRequestAdded represents a StakeRequestAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestAdded struct {
	RequestId *big.Int
	PublicKey common.Hash
	NodeID    string
	Amount    *big.Int
	StartTime *big.Int
	EndTime   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStakeRequestAdded is a free log retrieval operation binding the contract event 0x18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594.
//
// Solidity: event StakeRequestAdded(uint256 requestId, bytes indexed publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterStakeRequestAdded(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorStakeRequestAddedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "StakeRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorStakeRequestAddedIterator{contract: _MpcCoordinator.contract, event: "StakeRequestAdded", logs: logs, sub: sub}, nil
}

// WatchStakeRequestAdded is a free log subscription operation binding the contract event 0x18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594.
//
// Solidity: event StakeRequestAdded(uint256 requestId, bytes indexed publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchStakeRequestAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorStakeRequestAdded, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "StakeRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorStakeRequestAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeRequestAdded is a log parse operation binding the contract event 0x18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594.
//
// Solidity: event StakeRequestAdded(uint256 requestId, bytes indexed publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseStakeRequestAdded(log types.Log) (*MpcCoordinatorStakeRequestAdded, error) {
	event := new(MpcCoordinatorStakeRequestAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorStakeRequestStartedIterator is returned from FilterStakeRequestStarted and is used to iterate over the raw logs and unpacked data for StakeRequestStarted events raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestStartedIterator struct {
	Event *MpcCoordinatorStakeRequestStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorStakeRequestStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorStakeRequestStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorStakeRequestStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorStakeRequestStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorStakeRequestStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorStakeRequestStarted represents a StakeRequestStarted event raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestStarted struct {
	RequestId          *big.Int
	PublicKey          common.Hash
	ParticipantIndices []*big.Int
	NodeID             string
	Amount             *big.Int
	StartTime          *big.Int
	EndTime            *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterStakeRequestStarted is a free log retrieval operation binding the contract event 0x288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f30486.
//
// Solidity: event StakeRequestStarted(uint256 requestId, bytes indexed publicKey, uint256[] participantIndices, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterStakeRequestStarted(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorStakeRequestStartedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "StakeRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorStakeRequestStartedIterator{contract: _MpcCoordinator.contract, event: "StakeRequestStarted", logs: logs, sub: sub}, nil
}

// WatchStakeRequestStarted is a free log subscription operation binding the contract event 0x288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f30486.
//
// Solidity: event StakeRequestStarted(uint256 requestId, bytes indexed publicKey, uint256[] participantIndices, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchStakeRequestStarted(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorStakeRequestStarted, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "StakeRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorStakeRequestStarted)
				if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeRequestStarted is a log parse operation binding the contract event 0x288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f30486.
//
// Solidity: event StakeRequestStarted(uint256 requestId, bytes indexed publicKey, uint256[] participantIndices, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseStakeRequestStarted(log types.Log) (*MpcCoordinatorStakeRequestStarted, error) {
	event := new(MpcCoordinatorStakeRequestStarted)
	if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
