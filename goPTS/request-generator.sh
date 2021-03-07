#!/bin/bash

GET(){
	[ $# -ne 1 ] && echo "Please provide TestUID." && exit

	uid=$1

	curl -s -H 'Content-Type: application/json' -d '{"testuid":"'$uid'"}' http://localhost:8300/getStatus
}

SET(){
	myip=1.2.3.4
	
	rand=$((RANDOM%5))
	
	if [ $# -eq 3 ]
	then
		testtype=STRESS
		grpSize=3
		uid=$2
		testname=TEST-$rand
		testargs="ARGS-$rand"
		tstatus=$3
		myip=$1
	else
		echo "Please input <IP> <UID> <STATUS>"
		exit
	fi

	echo -n "[ $$ ] SET : $myip $uid $tstatus => "
	
	curl -s -H 'Content-Type: application/json' -d '{"testtype":"'$testtype'","groupsize":'$grpSize',"testuid":"'$uid'","test":"TEST-'$rand'","args":"ARGS-'$rand'","status":"'$tstatus'","myip":"'$myip'"}' http://localhost:8300/setStatus
}


validation_test(){
	while :
	do
		if [[ $((RANDOM%2)) -eq 0 ]]
		then
			set -x
			GET | tee GET.log 2>&1 &

			set +x
		else
			set -x
			SET | tee SET.log 2>&1 &
			set +x
		fi

		sleep $((RANDOM%7))
	done
}

waitForSync(){
	[ $# -ne 2 ] && echo "Required UID to proceed.!" && exit
	
	uid=$1
	sync_for=START

	SYNC_STATE=""

	st=$SECONDS

	while [[ "$SYNC_STATE" != "$sync_for" ]]
	do
		sleep 2
		SYNC_STATE=$(GET $uid)
		echo Synching for $sync_for...
		
		if [[ $((SECONDS-st)) -eq 120 ]]
		then
			echo Reached wait Timeout. Exiting wait loop.
			break
		fi
	done

	if [ "$SYNC_STATE" == "$sync_for" ]
	then
		echo "[SUT] Received: $SYNC_STATE" && return 0
	else
		return 1
	fi
}

simulate_pts_tp(){

	[ $# -lt 3 ] && echo "Please input <IP> <UID> <TES-STATUS>" && exit

	sleep $((RANDOM%5))

	ip=$1
	uid=$2
	test_status=$3

	SET $@ > cur-set.log

	waitForSync $uid START

	if [ $? == 0 ]
	then
	
			SET $ip $uid RUNNING
			
			echo  "[$(date +%H:%M:%S)] $MSG : STARTing My test."
			sleep $((RANDOM%7))

			SET $ip $uid DONE

			waitForSync $uid EXIT
	fi
}

for i in $(seq 1 3)
do
	simulate_pts_tp 1.2.3.$i 1 INIT > sim-$i.log 2>&1 &
done
