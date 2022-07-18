# MPC Controller
**This project is under active development**
## Project Components
We need to following components:
- A listener to web3 to add the task
- A manager to maintain the pending workflow tasks
- A poller to connect to MPC service
## Integration Tests
You can find test suites from `tests/Taskfile.yml`.
- Install library dependencies: `sudo apt install gcc-multilib libgmp-dev libssl-dev`
  (required on some Ubuntu OS).
- Install toolchain dependencies:  [Task](https://github.com/go-task/task), [Venom](https://github.com/ovh/venom),  [Foundry](https://github.com/foundry-rs/foundry), [Golang](https://go.dev/), [Rust](https://www.rust-lang.org/)
- Run tests: `task tests:testSuiteName`, e.g., `task tests:initiateStake`, `task tests:initiateStakeLoop`
- Check working directory: `cd $HOME/mpctest`
- Tear down: `task tests:cleanup`
## Todo List
- keystore to strength private key security
- automatic panic recover
- distributed trace, log and monitor
- deal with casual error: invalid nonce and nonce misused
- check and sync participant upon startup, there ere maybe groups created during mpc-controller downtime.
- add mpc-controller version info
- mechanism to check result from mpc-server and resume task on mpc-controller startup
- history even track for mpc-coordinator smart contract.
- log rotation with lumberjack: https://github.com/natefinch/lumberjack
- add main_test.go
- apply confluentinc/bincover: https://github.com/confluentinc/bincover
- restore data on startup
- automate tracking balance of addresses that receive principal and reward.
- take measures to deal with failed tasks
- take measures to avoid double-spend, maybe introduce SPE(single-participant-execution) strategy or consensus
- take measures to deal with package lost and disorder arrival
- ...
