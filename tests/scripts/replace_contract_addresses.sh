#!/usr/bin/env bash

OLD_AVALIDO_ADDRESS="0xec94c5fd10b9e200ed53ea6a563de4e450991966"
OLD_VALIDATOR_SELECTOR_ADDRESS="0xe48da35cddf27e1dfebe026e11723159da4c3c57"
OLD_ORACLE_ADDRESS="0x9caa7e123cc790a26e1edf24908b0dc0fdfaf492"
OLD_ORACLE_MANAGER_ADDRESS="0xad463d303f93b7ca14597c8abbf21954bbf31557"
OLD_MPC_MANAGER_ADDRESS="0x34aAe02798d6d18B4AD0c31bFbEc60fab9e3C5e4"

NEW_AVALIDO_ADDRESS="0xec94c5fd10b9e200ed53ea6a563de4e450991966"
NEW_VALIDATOR_SELECTOR_ADDRESS="0xe48da35cddf27e1dfebe026e11723159da4c3c57"
NEW_ORACLE_ADDRESS="0x9caa7e123cc790a26e1edf24908b0dc0fdfaf492"
NEW_ORACLE_MANAGER_ADDRESS="0xad463d303f93b7ca14597c8abbf21954bbf31557"
NEW_MPC_MANAGER_ADDRESS="0xd31b165d5816b4a97344197fb9a6c5c201ca690b"

# mpc-controller
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/docker/mpc-controller/configs/controller1.yaml
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/docker/mpc-controller/configs/controller2.yaml
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/docker/mpc-controller/configs/controller3.yaml
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/docker/mpc-controller/configs/controller4.yaml
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/docker/mpc-controller/configs/controller5.yaml
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/docker/mpc-controller/configs/controller6.yaml
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/docker/mpc-controller/configs/controller7.yaml

# oracle
sed -i "s|$OLD_ORACLE_ADDRESS|$NEW_ORACLE_ADDRESS|g" ./tests/docker/oracle/docker-compose.yml
sed -i "s|$OLD_ORACLE_MANAGER_ADDRESS|$NEW_ORACLE_MANAGER_ADDRESS|g" ./tests/docker/oracle/docker-compose.yml

# config file
sed -i "s|$OLD_AVALIDO_ADDRESS|$NEW_AVALIDO_ADDRESS|g" ./tests/scripts/config.sh
sed -i "s|$OLD_VALIDATOR_SELECTOR_ADDRESS|$NEW_ORACLE_ADDRESS|g" ./tests/scripts/config.sh
sed -i "s|$OLD_ORACLE_ADDRESS|$NEW_ORACLE_ADDRESS|g" ./tests/scripts/config.sh
sed -i "s|$OLD_ORACLE_MANAGER_ADDRESS|$NEW_ORACLE_MANAGER_ADDRESS|g" ./tests/scripts/config.sh
sed -i "s|$OLD_MPC_MANAGER_ADDRESS|$NEW_MPC_MANAGER_ADDRESS|g" ./tests/scripts/config.sh


