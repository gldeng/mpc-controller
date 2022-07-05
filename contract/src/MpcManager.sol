// SPDX-FileCopyrightText: 2022 Hyperelliptic Labs and RockX
// SPDX-License-Identifier: GPL-3.0
pragma solidity 0.8.15;

import "../lib/openzeppelin-contracts/contracts/security/Pausable.sol";
import "../lib/openzeppelin-contracts/contracts/security/ReentrancyGuard.sol";
import "../lib/openzeppelin-contracts/contracts/access/AccessControlEnumerable.sol";
import "./interfaces/IMpcManager.sol";
import "./interfaces/IMpcCoordinator.sol";

contract MpcManager is Pausable, ReentrancyGuard, AccessControlEnumerable, IMpcManager, IMpcCoordinator {
    // TODO:
    // Key these statements for observation and testing purposes only
    // Considering remove them later before everything fixed up and get into production mode.
    bytes public lastGenPubKey;
    address public lastGenAddress;

    enum RequestStatus {
        UNKNOWN,
        STARTED,
        COMPLETED
    }
    struct Request {
        bytes publicKey;
        bytes message;
        uint256[] participantIndices;
        RequestStatus status;
    }
    struct StakeRequestDetails {
        string nodeID;
        uint256 amount;
        uint256 startTime;
        uint256 endTime;
    }

    address private _avaLidoAddress;
    // groupId -> number of participants in the group
    mapping(bytes32 => uint256) private _groupParticipantCount;
    // groupId -> threshold
    mapping(bytes32 => uint256) private _groupThreshold;
    // groupId -> index -> participant
    mapping(bytes32 => mapping(uint256 => bytes)) private _groupParticipants;

    // key -> groupId
    mapping(bytes => KeyInfo) private _generatedKeys;

    // key -> index -> confirmed
    mapping(bytes => mapping(uint256 => bool)) private _keyConfirmations;

    // request status
    mapping(uint256 => Request) private _requests;
    mapping(uint256 => StakeRequestDetails) private _stakeRequestDetails;
    uint256 private _lastRequestId;

    // rewardedStakeTxId -> reportRewardedStakeCount
    mapping(bytes32 => uint256) private _reportRewardedStakeCounts;
    // rewardedStakeTxId -> joinExportRewardParticipantIndices
    mapping(bytes32 => uint256[]) private _joinExportRewardParticipantIndices;

    event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index);
    event KeyGenerated(bytes32 indexed groupId, bytes publicKey);
    event KeygenRequestAdded(bytes32 indexed groupId);
    event StakeRequestAdded(
        uint256 requestId,
        bytes indexed publicKey,
        string nodeID,
        uint256 amount,
        uint256 startTime,
        uint256 endTime
    );
    event StakeRequestStarted(
        uint256 requestId,
        bytes indexed publicKey,
        uint256[] participantIndices,
        string nodeID,
        uint256 amount,
        uint256 startTime,
        uint256 endTime
    );
    event SignRequestAdded(uint256 requestId, bytes indexed publicKey, bytes message);
    event SignRequestStarted(uint256 requestId, bytes indexed publicKey, bytes message);
    event ExportRewardRequestAdded(bytes32 indexed rewaredStakeTxId);
    event ExportRewardRequestStarted(bytes32 indexed rewaredStakeTxId, uint256[] participantIndices);

    constructor() {
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
    }

    // -------------------------------------------------------------------------
    //  External functions
    // -------------------------------------------------------------------------

    /**
     * @notice Send AVAX and start a StakeRequest.
     * @dev The received token will be immediately forwarded the the last generated MPC wallet
     * and the group members will handle the stake flow from the c-chain to the p-chain.
     */
    function requestStake(
        string calldata nodeID,
        uint256 amount,
        uint256 startTime,
        uint256 endTime
    ) external payable onlyAvaLido {
        require(lastGenAddress != address(0), "Key has not been generated yet.");
        require(msg.value == amount, "Incorrect value.");
        payable(lastGenAddress).transfer(amount);
        _handleStakeRequest(lastGenPubKey, nodeID, amount, startTime, endTime);
    }

    /**
     * @notice Admin will call this function to create an MPC group consisting of n members
     * and a specified threshold t. The signing can be performed by any t + 1 participants
     * from the group.
     * @param publicKeys The public keys which identify the n group members.
     * @param threshold The threshold t. Note: t + 1 participants are required to complete a
     * signing.
     */
    function createGroup(bytes[] calldata publicKeys, uint256 threshold) external onlyAdmin {
        // TODO: Refine ACL
        // TODO: Check public keys are valid
        require(publicKeys.length > 1, "A group requires 2 or more participants.");
        require(threshold >= 1 && threshold < publicKeys.length, "Invalid threshold");

        bytes memory b = bytes.concat(bytes32(threshold));
        for (uint256 i = 0; i < publicKeys.length; i++) {
            b = bytes.concat(b, publicKeys[i]);
        }
        bytes32 groupId = keccak256(b);

        uint256 count = _groupParticipantCount[groupId];
        require(count == 0, "Group already exists.");
        _groupParticipantCount[groupId] = publicKeys.length;
        _groupThreshold[groupId] = threshold;

        for (uint256 i = 0; i < publicKeys.length; i++) {
            _groupParticipants[groupId][i + 1] = publicKeys[i]; // Participant index is 1-based.
            emit ParticipantAdded(publicKeys[i], groupId, i + 1);
        }
    }

    /**
     * @notice Admin will call this function to tell the group members to generate a key. Multiple
     * keys can be generated for the same group.
     * @param groupId The id of the group which is deterministically derived from the public keys
     * of the ordered group members and the threshold.
     */
    function requestKeygen(bytes32 groupId) external onlyAdmin {
        // TODO: Refine ACL
        emit KeygenRequestAdded(groupId);
    }

    /**
     * @notice All group members have to report the generated key which also serves as the proof.
     * @param groupId The id of the mpc group.
     * @param myIndex The index of the participant in the group. This is 1-based.
     * @param generatedPublicKey The generated public key.
     */
    function reportGeneratedKey(
        bytes32 groupId,
        uint256 myIndex,
        bytes calldata generatedPublicKey
    ) external onlyGroupMember(groupId, myIndex) {
        KeyInfo storage info = _generatedKeys[generatedPublicKey];

        require(!info.confirmed, "Key has already been confirmed by all participants.");

        // TODO: Check public key valid
        _keyConfirmations[generatedPublicKey][myIndex] = true;

        if (_generatedKeyConfirmedByAll(groupId, generatedPublicKey)) {
            info.groupId = groupId;
            info.confirmed = true;
            // TODO: The two sentence below for naive testing purpose, to deal with them furher.
            lastGenPubKey = generatedPublicKey;
            lastGenAddress = _calculateAddress(generatedPublicKey);
            emit KeyGenerated(groupId, generatedPublicKey);
        }

        // TODO: Removed _keyConfirmations data after all confirmed
    }

    /**
     * @notice This is the primitive signing request. It may not be used in actual production.
     * @param publicKey The publicKey used for signing.
     * @param message An arbitrary message to be signed.
     */
    function requestSign(bytes calldata publicKey, bytes calldata message) external onlyAvaLido {
        KeyInfo memory info = _generatedKeys[publicKey];
        require(info.confirmed, "Key doesn't exist or has not been confirmed.");
        uint256 requestId = _getNextRequestId();
        Request storage status = _requests[requestId];
        status.publicKey = publicKey;
        status.message = message;
        emit SignRequestAdded(requestId, publicKey, message);
    }

    /**
     * @notice Participant has to call this function to join an MPC request. Each request
     * requires exactly t + 1 members to join.
     */
    function joinRequest(uint256 requestId, uint256 myIndex) external {
        // TODO: Add auth

        Request storage status = _requests[requestId];
        require(status.publicKey.length > 0, "Request doesn't exist.");

        KeyInfo memory info = _generatedKeys[status.publicKey];
        require(info.confirmed, "Public key doesn't exist or has not been confirmed.");

        uint256 threshold = _groupThreshold[info.groupId];
        require(status.participantIndices.length <= threshold, "Cannot join anymore.");

        _ensureSenderIsClaimedParticipant(info.groupId, myIndex);

        for (uint256 i = 0; i < status.participantIndices.length; i++) {
            require(status.participantIndices[i] != myIndex, "Already joined.");
        }
        status.participantIndices.push(myIndex);

        if (status.participantIndices.length == threshold + 1) {
            if (status.message.length > 0) {
                emit SignRequestStarted(requestId, status.publicKey, status.message);
            } else {
                StakeRequestDetails memory details = _stakeRequestDetails[requestId];
                if (details.amount > 0) {
                    emit StakeRequestStarted(
                        requestId,
                        status.publicKey,
                        status.participantIndices,
                        details.nodeID,
                        details.amount,
                        details.startTime,
                        details.endTime
                    );
                }
            }
        }
    }

    // -------------------------------------------------------------------------
    //  Admin functions
    // -------------------------------------------------------------------------

    function setAvaLidoAddress(address avaLidoAddress) external onlyAdmin {
        _avaLidoAddress = avaLidoAddress;
    }

    // -------------------------------------------------------------------------
    //  External view functions
    // -------------------------------------------------------------------------

    function getGroup(bytes32 groupId) external view returns (bytes[] memory participants, uint256 threshold) {
        uint256 count = _groupParticipantCount[groupId];
        require(count > 0, "Group doesn't exist.");
        bytes[] memory participants = new bytes[](count);
        threshold = _groupThreshold[groupId];

        for (uint256 i = 0; i < count; i++) {
            participants[i] = _groupParticipants[groupId][i + 1];
        }
        return (participants, threshold);
    }

    function getKey(bytes calldata publicKey) external view returns (KeyInfo memory keyInfo) {
        keyInfo = _generatedKeys[publicKey];
    }

    // -------------------------------------------------------------------------
    //  Modifiers
    // -------------------------------------------------------------------------

    modifier onlyAdmin() {
        // TODO: Define proper RBAC. For now just use deployer as admin.
        require(hasRole(DEFAULT_ADMIN_ROLE, msg.sender), "Caller is not admin.");
        _;
    }

    modifier onlyAvaLido() {
        require(msg.sender == _avaLidoAddress, "Caller is not AvaLido.");
        _;
    }

    modifier onlyGroupMember(bytes32 groupId, uint256 index) {
        _ensureSenderIsClaimedParticipant(groupId, index);
        _;
    }

    // -------------------------------------------------------------------------
    //  Internal functions
    // -------------------------------------------------------------------------

    // TODO: to deal with publickey param type modifier, currently use memory for testing convinience.
    function _handleStakeRequest(
        bytes memory publicKey,
        string calldata nodeID,
        uint256 amount,
        uint256 startTime,
        uint256 endTime
    ) internal {
        KeyInfo memory info = _generatedKeys[publicKey];
        require(info.confirmed, "Key doesn't exist or has not been confirmed.");

        // TODO: Validate input

        uint256 requestId = _getNextRequestId();
        Request storage status = _requests[requestId];
        status.publicKey = publicKey;
        // status.message is intentionally not set to indicate it's a StakeRequest

        StakeRequestDetails storage details = _stakeRequestDetails[requestId];

        details.nodeID = nodeID;
        details.amount = amount;
        details.startTime = startTime;
        details.endTime = endTime;
        emit StakeRequestAdded(requestId, publicKey, nodeID, amount, startTime, endTime);
    }

    function _getNextRequestId() internal returns (uint256) {
        _lastRequestId += 1;
        return _lastRequestId;
    }

    // -------------------------------------------------------------------------
    //  Private functions
    // -------------------------------------------------------------------------

    function _generatedKeyConfirmedByAll(bytes32 groupId, bytes calldata generatedPublicKey)
        private
        view
        returns (bool)
    {
        uint256 count = _groupParticipantCount[groupId];

        for (uint256 i = 0; i < count; i++) {
            if (!_keyConfirmations[generatedPublicKey][i + 1]) return false;
        }
        return true;
    }

    function _calculateAddress(bytes memory pub) private pure returns (address addr) {
        bytes32 hash = keccak256(pub);
        assembly {
            mstore(0, hash)
            addr := mload(0)
        }
    }

    function _ensureSenderIsClaimedParticipant(bytes32 groupId, uint256 index) private view {
        bytes memory publicKey = _groupParticipants[groupId][index];
        require(publicKey.length > 0, "Invalid groupId or index.");

        address member = _calculateAddress(publicKey);

        require(msg.sender == member, "Caller is not a group member");
    }

    // -------------------------------------------------------------------------
    //  Reward functions
    // -------------------------------------------------------------------------

    function reportRewardedStake(
        bytes32 groupId,
        uint256 myIndex,
        bytes32 txID
    ) external onlyGroupMember(groupId, myIndex) {
        uint256 groupMembers = 3; // todo: compare with number of group members.
        if (_reportRewardedStakeCounts[txID] < groupMembers) {
            _reportRewardedStakeCounts[txID] = _reportRewardedStakeCounts[txID]+1;
            if (_reportRewardedStakeCounts[txID] == groupMembers) {
                emit ExportRewardRequestAdded(txID);
            }
        }
    }

    function joinExportReward(
        bytes32 groupId,
        uint256 myIndex,
        bytes32 txID
    ) external onlyGroupMember(groupId, myIndex) {
        uint256 threshold = 1; // todo: compare with group threshold
        if (_joinExportRewardParticipantIndices[txID].length < threshold+1) {
            _joinExportRewardParticipantIndices[txID].push(myIndex);
            if (_joinExportRewardParticipantIndices[txID].length = threshold+1) {
                emit ExportRewardRequestStarted(txID, _joinExportRewardParticipantIndices[txID]);
            }
        }
    }
}
