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
- Run tests: `task tests:testSuiteName`, e.g., `task tests:initiateStake`, `task tests:initiateStakeLoop`.
- Check working directory: `cd $HOME/mpctest`. You can watch and analyze your log files from a terminal with [Inav](https://lnav.org/), shortcuts: `task lnavAvalanche`, `task lnavController`, `task lnavServer`, `task lnavServerMock`.
- Tear down: `task tests:cleanup`
## Monitoring
We have started using [Prometheus](https://prometheus.io/) for monitoring.
- From `prom` directory you can see what extra metrics supported.
- From `tests/configs/prom` you can see the Prometheus config for testings.
- Run `task prom` command to start Prometheus container for testings.
- Check [Prometheus Targets](http://localhost:9090/targets) 
- Check [Promethus Metrics](http://localhost:9090/metrics)
- Check [Promethus Graph](http://localhost:9090/graph) 