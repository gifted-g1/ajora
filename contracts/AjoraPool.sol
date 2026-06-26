
// contracts/AjoraPool.sol
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

contract AjoraPool is ReentrancyGuard, Pausable, Ownable {
    using EnumerableSet for EnumerableSet.AddressSet;

    struct PoolConfig {
        uint256 contributionAmount;
        uint256 totalSlots;
        uint256 totalRounds;
        uint256 contributionFrequency;
        uint256 interestRate;
        uint256 startTime;
        uint256 endTime;
    }

    struct Member {
        uint256 totalContributed;
        uint256 totalPayouts;
        uint256 payoutRound;
        uint256 penaltyPoints;
        bool isActive;
        bool hasReceivedPayout;
    }

    PoolConfig public config;
    EnumerableSet.AddressSet private members;
    mapping(address => Member) public membersInfo;
    mapping(uint256 => mapping(address => uint256)) public roundContributions;
    mapping(address => uint256[]) public contributionHistory;
    mapping(uint256 => address) public roundPayoutReceiver;
    mapping(uint256 => bool) public roundProcessed;

    uint256 public currentRound;
    bool public isComplete;
    address public governanceToken;

    event PoolCreated(address indexed creator, uint256 contributionAmount, uint256 totalSlots);
    event MemberJoined(address indexed member, uint256 slotNumber);
    event ContributionMade(address indexed member, uint256 round, uint256 amount);
    event PayoutProcessed(address indexed receiver, uint256 round, uint256 amount);
    event MemberDropped(address indexed member);
    event PoolCompleted(address indexed creator);

    constructor(
        uint256 _contributionAmount,
        uint256 _totalSlots,
        uint256 _totalRounds,
        uint256 _contributionFrequency,
        uint256 _interestRate,
        uint256 _startTime,
        uint256 _endTime,
        address _governanceToken
    ) {
        require(_totalSlots > 0 && _totalSlots <= 1000, "Invalid slots");
        require(_totalRounds > 0 && _totalRounds <= 52, "Invalid rounds");
        require(_contributionAmount > 0, "Invalid contribution");
        require(_startTime > block.timestamp, "Invalid start time");
        require(_endTime > _startTime, "Invalid end time");

        config = PoolConfig({
            contributionAmount: _contributionAmount,
            totalSlots: _totalSlots,
            totalRounds: _totalRounds,
            contributionFrequency: _contributionFrequency,
            interestRate: _interestRate,
            startTime: _startTime,
            endTime: _endTime
        });

        governanceToken = _governanceToken;
        currentRound = 1;

        emit PoolCreated(msg.sender, _contributionAmount, _totalSlots);
    }

    function joinPool() external payable whenNotPaused {
        require(block.timestamp < config.startTime, "Pool started");
        require(members.length() < config.totalSlots, "Pool full");
        require(!members.contains(msg.sender), "Already joined");

        if (config.contributionAmount > 0) {
            require(msg.value == config.contributionAmount, "Invalid contribution amount");
        }

        members.add(msg.sender);
        membersInfo[msg.sender] = Member({
            totalContributed: 0,
            totalPayouts: 0,
            payoutRound: 0,
            penaltyPoints: 0,
            isActive: true,
            hasReceivedPayout: false
        });

        emit MemberJoined(msg.sender, members.length());
    }

    function contribute() external payable whenNotPaused nonReentrant {
        require(block.timestamp >= config.startTime, "Pool not started");
        require(block.timestamp <= config.endTime, "Pool ended");
        require(members.contains(msg.sender), "Not a member");
        require(membersInfo[msg.sender].isActive, "Member inactive");
        require(!isComplete, "Pool completed");

        uint256 deadline = config.startTime + (currentRound - 1) * config.contributionFrequency;
        require(block.timestamp <= deadline + 7 days, "Round deadline passed");

        require(msg.value == config.contributionAmount, "Invalid contribution amount");

        membersInfo[msg.sender].totalContributed += msg.value;
        roundContributions[currentRound][msg.sender] += msg.value;
        contributionHistory[msg.sender].push(block.timestamp);

        emit ContributionMade(msg.sender, currentRound, msg.value);
    }

    function processPayout() external whenNotPaused nonReentrant onlyOwner {
        require(!isComplete, "Pool completed");
        require(block.timestamp >= config.endTime || currentRound > config.totalRounds, "Not complete");

        address receiver = getNextPayoutReceiver();
        require(receiver != address(0), "No receiver available");

        uint256 totalPool = address(this).balance;
        uint256 payoutAmount = (totalPool * (10000 + config.interestRate)) / 10000;

        require(payoutAmount > 0, "Insufficient funds");

        membersInfo[receiver].totalPayouts += payoutAmount;
        membersInfo[receiver].hasReceivedPayout = true;
        roundPayoutReceiver[currentRound] = receiver;
        roundProcessed[currentRound] = true;

        (bool success, ) = receiver.call{value: payoutAmount}("");
        require(success, "Payout failed");

        emit PayoutProcessed(receiver, currentRound, payoutAmount);

        currentRound++;

        if (currentRound > config.totalRounds) {
            isComplete = true;
            emit PoolCompleted(msg.sender);
        }
    }

    function getNextPayoutReceiver() internal view returns (address) {
        address[] memory memberList = members.values();
        uint256 startIndex = (currentRound - 1) % memberList.length;

        for (uint256 i = 0; i < memberList.length; i++) {
            uint256 index = (startIndex + i) % memberList.length;
            address candidate = memberList[index];
            if (membersInfo[candidate].isActive && 
                !membersInfo[candidate].hasReceivedPayout) {
                return candidate;
            }
        }
        return address(0);
    }

    function dropMember(address member) external whenNotPaused onlyOwner {
        require(members.contains(member), "Member not found");
        require(membersInfo[member].isActive, "Already inactive");

        membersInfo[member].isActive = false;
        membersInfo[member].penaltyPoints += 1;

        emit MemberDropped(member);
    }

    function getMemberCount() external view returns (uint256) {
        return members.length();
    }

    function getMembers() external view returns (address[] memory) {
        return members.values();
    }

    function getPoolStatus() external view returns (
        uint256 totalContributions,
        uint256 totalMembers,
        uint256 activeMembers,
        uint256 poolBalance
    ) {
        totalContributions = address(this).balance;
        totalMembers = members.length();
        activeMembers = 0;

        address[] memory memberList = members.values();
        for (uint256 i = 0; i < memberList.length; i++) {
            if (membersInfo[memberList[i]].isActive) {
                activeMembers++;
            }
        }

        poolBalance = address(this).balance;
    }

    receive() external payable {}
}

