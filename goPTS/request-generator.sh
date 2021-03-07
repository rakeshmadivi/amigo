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

	echo "[ $$ ] SET STATUS"
	
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

simulate_pts_tp(){

	[ $# -lt 3 ] && echo "Please input <IP> <UID> <TES-STATUS>" && exit

	ip=$1
	uid=$2
	test_status=$3

	SET $@ > cur-set.log
	
	while : 
	do
		STATUS=$(GET $uid 2>simulation-get.log | tr -d '"')

		MSG="$ip $uid"

		if [ "$STATUS" == "START" ]
		then
			SET $ip $uid RUNNING
			
			echo  "($MSG) STARTing My test."
			sleep 7

			SET $ip $uid DONE

		elif [ "$STATUS" == "RUNNING" ]
		then
			echo "($MSG) Test are already running..."

		elif [ "$STATUS" == "EXIT" ]
		then
			echo EXITing...
			exit
		fi

		echo "[ $$ ] ($MSG - $STATUS) syncing...."

		sleep 3
	done
}

simulate_pts_tp 1.2.3.4 1 INIT &
simulate_pts_tp 1.2.3.5 1 INIT &
simulate_pts_tp 1.2.3.6 1 INIT &
