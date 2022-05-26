pragma solidity 0.8.14;

uint256 constant STAKE_PERIOD = 14 days;
string constant NODE_ID = "NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5";

interface IMpcManagerSimple {
    function requestStake(string calldata nodeID, uint256 amount, uint256 startTime, uint256 endTime) external payable;
}

// This version of AvaLido contract is simplified for testing of MPC-Manager stake feature.
contract AvaLido {
    address public mpcManagerAddress_;
    IMpcManagerSimple public mpcManager;

    constructor(
        address mpcManagerAddress
    ) payable {
        mpcManagerAddress_ = mpcManagerAddress;
        mpcManager = IMpcManagerSimple(mpcManagerAddress);
    }

    receive() payable external {
    }

    function getBalance() public view returns (uint256) {
        return address(this).balance;
    }

    function initiateStake(uint256 amount) external returns (uint256) {
        payable(mpcManagerAddress_).transfer(amount);

        uint256 startTime = block.timestamp + 30 seconds;
        uint256 endTime = startTime + STAKE_PERIOD;
        mpcManager.requestStake(NODE_ID, amount, startTime, endTime);

        return amount;
    }
}
