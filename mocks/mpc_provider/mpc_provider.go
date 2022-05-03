package mpc_provider

import (
	"context"
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/token"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/juju/errors"
	"math/big"
)

const (
	MinimumToEnsureBalance = 1_000_000_000_000_000_000
	Bytecode               = `608060405234801561001057600080fd5b50612606806100206000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063b567d4ba1161005b578063b567d4ba14610111578063dd6bd14914610142578063e661d90d1461015e578063fae3a93c1461017a57610088565b80632dbf03441461008d5780632ed92550146100a95780632f7e3d17146100c55780637fed84f2146100e1575b600080fd5b6100a760048036038101906100a291906115f9565b610196565b005b6100c360048036038101906100be91906116a9565b610311565b005b6100df60048036038101906100da9190611584565b610747565b005b6100fb60048036038101906100f6919061153f565b610886565b6040516101089190611eaf565b60405180910390f35b61012b600480360381019061012691906114aa565b6108e6565b604051610139929190611cb2565b60405180910390f35b61015c60048036038101906101579190611452565b610af7565b005b610178600480360381019061017391906114aa565b610e38565b005b610194600480360381019061018f91906114d3565b610e68565b005b6000600388886040516101aa929190611c5c565b9081526020016040518091039020604051806040016040529081600082015481526020016001820160009054906101000a900460ff1615151515815250509050806020015161022e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161022590611d2f565b60405180910390fd5b6000610238610fae565b905060006005600083815260200190815260200160002090508989826000019190610264929190611202565b506000600660008481526020019081526020016000209050888882600001919061028f929190611288565b508681600101819055508581600201819055508481600301819055508a8a6040516102bb929190611c5c565b60405180910390207f18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594848b8b8b8b8b6040516102fc96959493929190611f9b565b60405180910390a25050505050505050505050565b60006005600084815260200190815260200160002090506000816000018054610339906121f0565b90501161037b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161037290611e2f565b60405180910390fd5b60006003826000016040516103909190611c9b565b9081526020016040518091039020604051806040016040529081600082015481526020016001820160009054906101000a900460ff16151515158152505090508060200151610414576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161040b90611d6f565b60405180910390fd5b60006001600083600001518152602001908152602001600020549050808360020180549050111561047a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161047190611e0f565b60405180910390fd5b610488826000015185610fd2565b60005b836002018054905081101561053657848460020182815481106104d7577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b90600052602060002001541415610523576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161051a90611e8f565b60405180910390fd5b808061052e9061223c565b91505061048b565b508260020184908060018154018082558091505060019003906000526020600020016000909190919091505560018161056f919061212e565b8360020180549050141561074057600083600101805461058e906121f0565b905011156105f057826000016040516105a79190611c9b565b60405180910390207f279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad30386856001016040516105e3929190611f6b565b60405180910390a261073f565b600060066000878152602001908152602001600020604051806080016040529081600082018054610620906121f0565b80601f016020809104026020016040519081016040528092919081815260200182805461064c906121f0565b80156106995780601f1061066e57610100808354040283529160200191610699565b820191906000526020600020905b81548152906001019060200180831161067c57829003601f168201915b505050505081526020016001820154815260200160028201548152602001600382015481525050905060008160200151111561073d57836000016040516106e09190611c9b565b60405180910390207f288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f304868786600201846000015185602001518660400151876060015160405161073496959493929190611eca565b60405180910390a25b505b5b5050505050565b60006003858560405161075b929190611c5c565b9081526020016040518091039020604051806040016040529081600082015481526020016001820160009054906101000a900460ff161515151581525050905080602001516107df576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107d690611d2f565b60405180910390fd5b60006107e9610fae565b905060006005600083815260200190815260200160002090508686826000019190610815929190611202565b508484826001019190610829929190611202565b50868660405161083a929190611c5c565b60405180910390207ffd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b2262883878760405161087593929190611f39565b60405180910390a250505050505050565b61088e61130e565b600383836040516108a0929190611c5c565b9081526020016040518091039020604051806040016040529081600082015481526020016001820160009054906101000a900460ff161515151581525050905092915050565b606060008060008085815260200190815260200160002054905060008111610943576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161093a90611d8f565b60405180910390fd5b60008167ffffffffffffffff811115610985577f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6040519080825280602002602001820160405280156109b857816020015b60608152602001906001900390816109a35790505b5090506001600086815260200190815260200160002054925060005b82811015610aec576002600087815260200190815260200160002060006001836109fe919061212e565b81526020019081526020016000208054610a17906121f0565b80601f0160208091040260200160405190810160405280929190818152602001828054610a43906121f0565b8015610a905780601f10610a6557610100808354040283529160200191610a90565b820191906000526020600020905b815481529060010190602001808311610a7357829003601f168201915b5050505050828281518110610ace577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60200260200101819052508080610ae49061223c565b9150506109d4565b508093505050915091565b60018383905011610b3d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b3490611d4f565b60405180910390fd5b60018110158015610b5057508282905081105b610b8f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b8690611dcf565b60405180910390fd5b60008160001b604051602001610ba59190611c41565b604051602081830303815290604052905060005b84849050811015610c465781858583818110610bfe577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b9050602002810190610c109190611ff7565b604051602001610c2293929190611c75565b60405160208183030381529060405291508080610c3e9061223c565b915050610bb9565b50600081805190602001209050600080600083815260200190815260200160002054905060008114610cad576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ca490611e6f565b60405180910390fd5b858590506000808481526020019081526020016000208190555083600160008481526020019081526020016000208190555060005b86869050811015610e2f57868682818110610d26577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b9050602002810190610d389190611ff7565b600260008681526020019081526020016000206000600185610d5a919061212e565b81526020019081526020016000209190610d75929190611202565b50868682818110610daf577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b9050602002810190610dc19190611ff7565b604051610dcf929190611c5c565b60405180910390207f39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a84600184610e06919061212e565b604051610e14929190611ce2565b60405180910390a28080610e279061223c565b915050610ce2565b50505050505050565b807f5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa732760405160405180910390a250565b8383610e748282610fd2565b600060038585604051610e88929190611c5c565b908152602001604051809103902090508060010160009054906101000a900460ff1615610eea576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ee190611def565b60405180910390fd5b600160048686604051610efe929190611c5c565b9081526020016040518091039020600088815260200190815260200160002060006101000a81548160ff021916908315150217905550610f3f878686611148565b15610fa55786816000018190555060018160010160006101000a81548160ff021916908315150217905550867f767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d8686604051610f9c929190611d0b565b60405180910390a25b50505050505050565b6000600160076000828254610fc3919061212e565b92505081905550600754905090565b60006002600084815260200190815260200160002060008381526020019081526020016000208054611003906121f0565b80601f016020809104026020016040519081016040528092919081815260200182805461102f906121f0565b801561107c5780601f106110515761010080835404028352916020019161107c565b820191906000526020600020905b81548152906001019060200180831161105f57829003601f168201915b5050505050905060008151116110c7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110be90611e4f565b60405180910390fd5b60006110d2826111e6565b90508073ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614611142576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161113990611daf565b60405180910390fd5b50505050565b60008060008086815260200190815260200160002054905060005b818110156111d8576004858560405161117d929190611c5c565b9081526020016040518091039020600060018361119a919061212e565b815260200190815260200160002060009054906101000a900460ff166111c5576000925050506111df565b80806111d09061223c565b915050611163565b5060019150505b9392505050565b6000808280519060200120905080600052600051915050919050565b82805461120e906121f0565b90600052602060002090601f0160209004810192826112305760008555611277565b82601f1061124957803560ff1916838001178555611277565b82800160010185558215611277579182015b8281111561127657823582559160200191906001019061125b565b5b509050611284919061132d565b5090565b828054611294906121f0565b90600052602060002090601f0160209004810192826112b657600085556112fd565b82601f106112cf57803560ff19168380011785556112fd565b828001600101855582156112fd579182015b828111156112fc5782358255916020019190600101906112e1565b5b50905061130a919061132d565b5090565b6040518060400160405280600080191681526020016000151581525090565b5b8082111561134657600081600090555060010161132e565b5090565b60008083601f84011261135c57600080fd5b8235905067ffffffffffffffff81111561137557600080fd5b60208301915083602082028301111561138d57600080fd5b9250929050565b6000813590506113a3816125a2565b92915050565b60008083601f8401126113bb57600080fd5b8235905067ffffffffffffffff8111156113d457600080fd5b6020830191508360018202830111156113ec57600080fd5b9250929050565b60008083601f84011261140557600080fd5b8235905067ffffffffffffffff81111561141e57600080fd5b60208301915083600182028301111561143657600080fd5b9250929050565b60008135905061144c816125b9565b92915050565b60008060006040848603121561146757600080fd5b600084013567ffffffffffffffff81111561148157600080fd5b61148d8682870161134a565b935093505060206114a08682870161143d565b9150509250925092565b6000602082840312156114bc57600080fd5b60006114ca84828501611394565b91505092915050565b600080600080606085870312156114e957600080fd5b60006114f787828801611394565b94505060206115088782880161143d565b935050604085013567ffffffffffffffff81111561152557600080fd5b611531878288016113a9565b925092505092959194509250565b6000806020838503121561155257600080fd5b600083013567ffffffffffffffff81111561156c57600080fd5b611578858286016113a9565b92509250509250929050565b6000806000806040858703121561159a57600080fd5b600085013567ffffffffffffffff8111156115b457600080fd5b6115c0878288016113a9565b9450945050602085013567ffffffffffffffff8111156115df57600080fd5b6115eb878288016113a9565b925092505092959194509250565b600080600080600080600060a0888a03121561161457600080fd5b600088013567ffffffffffffffff81111561162e57600080fd5b61163a8a828b016113a9565b9750975050602088013567ffffffffffffffff81111561165957600080fd5b6116658a828b016113f3565b955095505060406116788a828b0161143d565b93505060606116898a828b0161143d565b925050608061169a8a828b0161143d565b91505092959891949750929550565b600080604083850312156116bc57600080fd5b60006116ca8582860161143d565b92505060206116db8582860161143d565b9150509250929050565b60006116f18383611881565b905092915050565b60006117058383611c23565b60208301905092915050565b600061171c82612088565b61172681856120ce565b9350836020820285016117388561204e565b8060005b85811015611774578484038952815161175585826116e5565b9450611760836120b4565b925060208a0199505060018101905061173c565b50829750879550505050505092915050565b600061179182612093565b61179b81856120df565b93506117a68361205e565b8060005b838110156117de576117bb826122ed565b6117c588826116f9565b97506117d0836120c1565b9250506001810190506117aa565b5085935050505092915050565b6117f48161218e565b82525050565b6118038161219a565b82525050565b6118128161219a565b82525050565b6118296118248261219a565b612285565b82525050565b600061183b8385612101565b93506118488385846121ae565b61185183612300565b840190509392505050565b60006118688385612112565b93506118758385846121ae565b82840190509392505050565b600061188c8261209e565b61189681856120f0565b93506118a68185602086016121bd565b6118af81612300565b840191505092915050565b60006118c58261209e565b6118cf8185612112565b93506118df8185602086016121bd565b80840191505092915050565b600081546118f8816121f0565b6119028186612101565b9450600182166000811461191d576001811461192f57611962565b60ff1983168652602086019350611962565b61193885612073565b60005b8381101561195a5781548189015260018201915060208101905061193b565b808801955050505b50505092915050565b60008154611978816121f0565b6119828186612112565b9450600182166000811461199d57600181146119ae576119e1565b60ff198316865281860193506119e1565b6119b785612073565b60005b838110156119d9578154818901526001820191506020810190506119ba565b838801955050505b50505092915050565b60006119f6838561211d565b9350611a038385846121ae565b611a0c83612300565b840190509392505050565b6000611a22826120a9565b611a2c818561211d565b9350611a3c8185602086016121bd565b611a4581612300565b840191505092915050565b6000611a5d602c8361211d565b9150611a688261231e565b604082019050919050565b6000611a8060288361211d565b9150611a8b8261236d565b604082019050919050565b6000611aa360338361211d565b9150611aae826123bc565b604082019050919050565b6000611ac660148361211d565b9150611ad18261240b565b602082019050919050565b6000611ae9601c8361211d565b9150611af482612434565b602082019050919050565b6000611b0c60118361211d565b9150611b178261245d565b602082019050919050565b6000611b2f60338361211d565b9150611b3a82612486565b604082019050919050565b6000611b5260148361211d565b9150611b5d826124d5565b602082019050919050565b6000611b7560168361211d565b9150611b80826124fe565b602082019050919050565b6000611b9860198361211d565b9150611ba382612527565b602082019050919050565b6000611bbb60158361211d565b9150611bc682612550565b602082019050919050565b6000611bde600f8361211d565b9150611be982612579565b602082019050919050565b604082016000820151611c0a60008501826117fa565b506020820151611c1d60208501826117eb565b50505050565b611c2c816121a4565b82525050565b611c3b816121a4565b82525050565b6000611c4d8284611818565b60208201915081905092915050565b6000611c6982848661185c565b91508190509392505050565b6000611c8182866118ba565b9150611c8e82848661185c565b9150819050949350505050565b6000611ca7828461196b565b915081905092915050565b60006040820190508181036000830152611ccc8185611711565b9050611cdb6020830184611c32565b9392505050565b6000604082019050611cf76000830185611809565b611d046020830184611c32565b9392505050565b60006020820190508181036000830152611d2681848661182f565b90509392505050565b60006020820190508181036000830152611d4881611a50565b9050919050565b60006020820190508181036000830152611d6881611a73565b9050919050565b60006020820190508181036000830152611d8881611a96565b9050919050565b60006020820190508181036000830152611da881611ab9565b9050919050565b60006020820190508181036000830152611dc881611adc565b9050919050565b60006020820190508181036000830152611de881611aff565b9050919050565b60006020820190508181036000830152611e0881611b22565b9050919050565b60006020820190508181036000830152611e2881611b45565b9050919050565b60006020820190508181036000830152611e4881611b68565b9050919050565b60006020820190508181036000830152611e6881611b8b565b9050919050565b60006020820190508181036000830152611e8881611bae565b9050919050565b60006020820190508181036000830152611ea881611bd1565b9050919050565b6000604082019050611ec46000830184611bf4565b92915050565b600060c082019050611edf6000830189611c32565b8181036020830152611ef18188611786565b90508181036040830152611f058187611a17565b9050611f146060830186611c32565b611f216080830185611c32565b611f2e60a0830184611c32565b979650505050505050565b6000604082019050611f4e6000830186611c32565b8181036020830152611f6181848661182f565b9050949350505050565b6000604082019050611f806000830185611c32565b8181036020830152611f9281846118eb565b90509392505050565b600060a082019050611fb06000830189611c32565b8181036020830152611fc38187896119ea565b9050611fd26040830186611c32565b611fdf6060830185611c32565b611fec6080830184611c32565b979650505050505050565b6000808335600160200384360303811261201057600080fd5b80840192508235915067ffffffffffffffff82111561202e57600080fd5b60208301925060018202360383131561204657600080fd5b509250929050565b6000819050602082019050919050565b60008190508160005260206000209050919050565b60008190508160005260206000209050919050565b600081519050919050565b600081549050919050565b600081519050919050565b600081519050919050565b6000602082019050919050565b6000600182019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b600082825260208201905092915050565b600081905092915050565b600082825260208201905092915050565b6000612139826121a4565b9150612144836121a4565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156121795761217861228f565b5b828201905092915050565b6000819050919050565b60008115159050919050565b6000819050919050565b6000819050919050565b82818337600083830152505050565b60005b838110156121db5780820151818401526020810190506121c0565b838111156121ea576000848401525b50505050565b6000600282049050600182168061220857607f821691505b6020821081141561221c5761221b6122be565b5b50919050565b600061223561223083612311565b612184565b9050919050565b6000612247826121a4565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82141561227a5761227961228f565b5b600182019050919050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b60006122f98254612222565b9050919050565b6000601f19601f8301169050919050565b60008160001c9050919050565b7f4b657920646f65736e2774206578697374206f7220686173206e6f742062656560008201527f6e20636f6e6669726d65642e0000000000000000000000000000000000000000602082015250565b7f412067726f75702072657175697265732032206f72206d6f726520706172746960008201527f636970616e74732e000000000000000000000000000000000000000000000000602082015250565b7f5075626c6963206b657920646f65736e2774206578697374206f72206861732060008201527f6e6f74206265656e20636f6e6669726d65642e00000000000000000000000000602082015250565b7f47726f757020646f65736e27742065786973742e000000000000000000000000600082015250565b7f43616c6c6572206973206e6f7420612067726f7570206d656d62657200000000600082015250565b7f496e76616c6964207468726573686f6c64000000000000000000000000000000600082015250565b7f4b65792068617320616c7265616479206265656e20636f6e6669726d6564206260008201527f7920616c6c207061727469636970616e74732e00000000000000000000000000602082015250565b7f43616e6e6f74206a6f696e20616e796d6f72652e000000000000000000000000600082015250565b7f5265717565737420646f65736e27742065786973742e00000000000000000000600082015250565b7f496e76616c69642067726f75704964206f7220696e6465782e00000000000000600082015250565b7f47726f757020616c7265616479206578697374732e0000000000000000000000600082015250565b7f416c7265616479206a6f696e65642e0000000000000000000000000000000000600082015250565b6125ab8161219a565b81146125b657600080fd5b50565b6125c2816121a4565b81146125cd57600080fd5b5056fea26469706673582212204751e8a542833ea7bb6b67c4e466702f6482d90ed03d733a258bbe83801e6ec864736f6c63430008040033`
)

type MpcProvider struct {
	chainId         int64
	rpcClient       *ethclient.Client
	wsClient        *ethclient.Client
	RpcCoordinator  *contract.Coordinator // created automatically after calling DeployContract
	WsCoordinator   *contract.Coordinator // created automatically after calling DeployContract
	ContractAddress *common.Address       // created automatically after calling DeployContract
	privateKey      *ecdsa.PrivateKey
}

func New(chainId int64, privKey *ecdsa.PrivateKey, rpcClient, wsClient *ethclient.Client) *MpcProvider {
	return &MpcProvider{
		chainId:        chainId,
		rpcClient:      rpcClient,
		wsClient:       wsClient,
		RpcCoordinator: nil,
		WsCoordinator:  nil,
		privateKey:     privKey,
	}
}

// DeployContract deploy new coordinator smart contract instance and return its address
func (m *MpcProvider) DeployContract() (*common.Address, *types.Receipt, error) {
	addr, rcp, err := contract.Deploy(m.chainId, m.rpcClient, m.privateKey, Bytecode)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	rpcCoordinator, err := contract.NewCoordinator(m.chainId, addr, m.rpcClient)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	wsCoordinator, err := contract.NewCoordinator(m.chainId, addr, m.wsClient)
	if err != nil {
		return nil, nil, errors.Trace(err)
	}
	m.RpcCoordinator = rpcCoordinator
	m.WsCoordinator = wsCoordinator
	m.ContractAddress = addr

	logger.Info("Coordinator contract deployed", logger.Field{"contractAddress", addr.Hex()})
	return addr, rcp, nil
}

// CreateGroup creates group with coordinator smart contract and return the created group id
// todo: return receipt?
func (m *MpcProvider) CreateGroup(participantPubKeys []*ecdsa.PublicKey, threshold int64) (string, error) {
	if len(participantPubKeys) < 3 {
		return "", errors.New("Require at least three participants to create a group")
	}

	if m.RpcCoordinator == nil || m.WsCoordinator == nil {
		return "", errors.New("Nil coordinators provided")
	}

	participants := crypto.MarshalPubkeys(participantPubKeys)
	var participants_ [][]byte
	for _, participant := range participants {
		participants_ = append(participants_, participant[1:])
	}

	_, err := m.RpcCoordinator.CreateGroup_(m.privateKey, participants_, threshold)
	if err != nil {
		return "", errors.Trace(err)
	}

	groupId, err := m.waitForAllParticipantsAdded(participants_)
	if err != nil {
		return "", errors.Trace(err)
	}

	err = m.ensureBalance(crypto.PubkeysToAddresses(participantPubKeys))
	if err != nil {
		return "", errors.Trace(err)
	}

	logger.Info("Group created",
		logger.Field{"groupId", groupId},
		logger.Field{"contractAddress", m.ContractAddress.Hex()})
	return groupId, nil
}

func (m *MpcProvider) ensureBalance(participantAddrs []*common.Address) error {
	for _, addr := range participantAddrs {
		bal, err := m.rpcClient.BalanceAt(context.Background(), *addr, nil)
		if err != nil {
			return errors.Trace(err)
		}

		if bal.Cmp(big.NewInt(MinimumToEnsureBalance)) < 0 {
			err = token.TransferInCChain(m.rpcClient, m.chainId, m.privateKey, addr, MinimumToEnsureBalance)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

// todo: check whether the emit group id is fully corresponding to the given participant public keys.
func (m *MpcProvider) waitForAllParticipantsAdded(participantPubKeys [][]byte) (string, error) {
	events := make(chan *contract.MpcCoordinatorParticipantAdded)

	var start = uint64(1)
	opts := new(bind.WatchOpts)
	opts.Start = &start
	sub, err := m.WsCoordinator.WatchParticipantAdded(opts, events, participantPubKeys)
	if err != nil {
		return "", errors.Trace(err)
	}

	var listenErr error
	var groupIDHex string

listen:
	for {
		select {
		case err := <-sub.Err():
			listenErr = err
			break listen
		case evt := <-events:
			groupIDHex = common.Bytes2Hex(evt.GroupId[:])
			break listen
		}
	}

	sub.Unsubscribe()

	return groupIDHex, listenErr
}
