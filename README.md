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
