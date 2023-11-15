#!/bin/bash

set -e

echo "Cleaning Raven."
if sudo rvn destroy; then
	echo "Raven destroyed"
else
	exit 1
fi
if sudo rvn -v build; then
	echo "Built Raven Topology"
else
	exit 1
fi
if sudo rvn -v deploy; then
	echo "Deployed Raven Topology"
else
	exit 1
fi

nodes=`sudo rvn status 2>&1 | sed 's/.*\(msg=.*\)/\1/g' | grep "  " | cut -d " " -f 3`

echo "Pinging nodes until topology is ready."
if sudo rvn pingwait $nodes; then
	echo "Raven Topology UP"
else
	exit 1
fi
if sudo rvn status; then
	echo "Raven Status (generate ansible)"
else
	exit 1
fi
echo "Configuring Raven Topology."
if sudo rvn configure -v; then
	echo "Raven Status (generate ansible)"
else
	exit 1
fi

echo "Success."
