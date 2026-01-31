#!/bin/bash

sftp ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR} <<EOF
    rm ${REMOTE_DIR}/*-cli
		put "dist/benchmarker-cli_${REMOTE_OS}_${REMOTE_ARCH}/benchmarker-cli"
		put "dist/feeder-cli_${REMOTE_OS}_${REMOTE_ARCH}/feeder-cli"
		put config.yaml
		quit
	EOF