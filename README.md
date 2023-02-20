[![Docker Build workflow](https://github.com/AvaLido/mpc-controller/actions/workflows/build-docker.yml/badge.svg)](https://github.com/AvaLido/mpc-controller/actions?workflow=build-docker)
[![Lint workflow](https://github.com/AvaLido/mpc-controller/actions/workflows/lint.yml/badge.svg)](https://github.com/AvaLido/mpc-controller/actions?workflow=lint)
[![Test workflow](https://github.com/AvaLido/mpc-controller/actions/workflows/test.yml/badge.svg)](https://github.com/AvaLido/mpc-controller/actions?workflow=test)

# MPC Controller

## Architecture
![artchitecture](/docs/architecture.png)

This MPC controller is a component of the AvaLido liquid staking project which contains mainly three components:
1. A set of smart contract deployed on the Avalanche C-Chain. The two main smart contract instances are:
  * the main `AvaLido` contract which contains the main liquid staking logics. Users will deposit their AVAX to this
`AvaLido` contract and the contract will trigger the staking process under certain conditions.
  * the `MpcManager` contract which accepts request from the `AvaLido` contract and serves as a coordinator for the MPC
backend which execute the staking and other flows.
2. An MPC controller which listens to the events from the Avalanche network and executes the predefined flows.
3. An MPC server which provide keygen and signing services to the MPC controller. It talks to other peers of the same
MPC group to run the keygen and signing protocols.

This repo contains the code of the MPC controller component.

## Design of the MPC Controller
### Overview
![design](/docs/mpc-controller-design.png)

The main responsibility of the MPC Controller is to get the events from the Avalanche network and perform the predefined
flows. The `subscriber`, `indexer` and `synchronizer` are used to get the events for the Avalanche network. All the
events received will be added to a queue whereby later on the `router` will pick up the events one by one and pass them
to the respective handlers registered with it to respond to the events. The handlers will create subsequent tasks to
handle each event. The tasks will be added to a queue and will be served by the worker pool.

### Tasks
Task is at the core of this repo. Every task has its own state. The workers in the worker pool will pick up a task when
it's idle. It calls the `Next` function of the task with a `TaskContext` cached in the pool. The task may create new
other tasks, if so, the worker will receive the created tasks as a output of the call and submit them in the worker
pool. After every call to `Next`, the worker may either drop the task (if the task `IsDone` or `FailedPermanently`) or
re-queue the task. Some tasks can be run in parallel (e.g. UTXO based txns on P-chain) while some have to be run
sequentially (e.g. interacting with contract on C-Chain requiring consumption of EOA `nonce`). `IsSequential` can tell
if the task will be added in sequential queue or parallel queue.
```go
type Task interface {
	GetId() string
	Next(ctx TaskContext) ([]Task, error)
	IsDone() bool
	FailedPermanently() bool
	IsSequential() bool
}
```

Take note that tasks are composable. A task can be composed of other tasks. For example a `C2P` task (i.e. moving AVAX
from C-chain to P-chain) is composed of an `ExportFromPChain` task and an `ImportIntoCChain` task. It runs
the `ExportFromPChain` task first and creates an `ImportIntoCChain` task once the `ExportFromPChain` is done. It fails
if either task fails. It is sequential while the inner `ExportFromPChain` is running and becomes non-sequential (i.e.
parallelizable) when the inner `ImportIntoCChain` is running.


### Folder Structure
  - [core](core) contains the basic types and interfaces. It should not depend on any other packages in the repo.
  - [contract](contract) contains a generated go binding for the `MpcManager` contract
  - [eventhandlercontext](eventhandlercontext) is the implementation of the context required by the event handlers
  - [indexer](indexer) scans indexes `AddDelegatorTx` and `UTXO`s periodically.
  - [mpcclient](mpcclient) contains a gRPC client used to interact with the MPC server.
  - [pool](pool) is the worker pool implementation which contains a sequential worker and a parallel pool.
  - [prom](prom) defines the prometheus metrics.
  - [proto](proto) conains the protobuf generated code used by gRPC.
  - [router](router) registers the event handlers and processes the events using the registered handlers.
  - [storage](storage) contains the storage implementations.
  - [subscriber](subscriber) is the subscriber to the contract events from the C-chain.
  - [synchronizer](synchronizer) is used when the program starts up and needs to grab historic contract events (to
initialize the group and key information).
  - [taskcontext](taskcontext) is the `TaskContext` implementation.
  - [tasks](tasks) contains all tasks.

## Integration Tests
You can find test suites from `tests/Taskfile.yml`. 
- Install library dependencies: `sudo apt install gcc-multilib libgmp-dev libssl-dev`
  (required on some Ubuntu OS).
- Install toolchain dependencies:  [Task](https://github.com/go-task/task), [Venom](https://github.com/ovh/venom),  [Foundry](https://github.com/foundry-rs/foundry), [Golang](https://go.dev/), [Rust](https://www.rust-lang.org/)
- Run tests: `task tests:testSuiteName`, e.g., `task tests:initiateStake`, `task tests:initiateStakeLoop -- {{.LOOP_TIMES}}`. In case you need to rebuild before you repeat test, you can manually delete sub-directory under `$HOME/mpctest`. 
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
