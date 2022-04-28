address=$1
ip=localhost

pkill -f mpc-controller

sleep 2

./mpc-controller --rpc-url not-needed --mpc-url http://$ip:8003 --private-key 5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b --coordinator-address $address  > ./p3.log 2>&1 &
./mpc-controller --rpc-url not-needed --mpc-url http://$ip:8001 --private-key 59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21 --coordinator-address $address  > ./p1.log 2>&1 &
./mpc-controller --rpc-url not-needed --mpc-url http://$ip:8002 --private-key 6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33 --coordinator-address $address  > ./p2.log 2>&1 &

