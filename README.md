# MPC Controller
**This project is under active development**
## Project Components
We need to following components:
- A listener to web3 to add the task
- A manager to maintain the pending workflow tasks
- A poller to connect to MPC service
## Integration Tests
You can find test suites from `tests/testsuites` directory
- Library deps: `sudo apt install gcc-multilib libgmp-dev libssl-dev`
  (required in some Ubuntu OS.)
- Toolchain deps:  [Task](https://github.com/go-task/task), [Venom](https://github.com/ovh/venom),  [Foundry](https://github.com/foundry-rs/foundry), [Golang](https://go.dev/), [Rust](https://www.rust-lang.org/)
- Run tests: `task tests:testSuiteName`
- Check working directory: `cd /tmp/mpctest`
- Tear down: `task tests:cleanup`