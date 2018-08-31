#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

REPO_DIR="$( cd ${SCRIPT_DIR} && git rev-parse --show-toplevel )"
PACKAGE_NAME="$(basename "${REPO_DIR}")"

BIN_DIR=${REPO_DIR}/bin


printf "Location of run-build.sh = ${SCRIPT_DIR}\n"
printf "Location of bin dir = ${BIN_DIR}\n"
printf "Package name = ${PACKAGE_NAME}\n"

if [ `ls -1 ${REPO_DIR}/*go 2>/dev/null | wc -l ` -gt 0 ]
then
	printf "\n${REPO_DIR} is a GO project - ok\n"
else
	printf "\n${REPO_DIR} is a non-GO project - nok.\n Exiting....\n"
	exit
fi



mkdir -p ${BIN_DIR}

rm -rf ${BIN_DIR}/*





platforms=("windows/amd64" "darwin/amd64" "linux/amd64")

for platform in "${platforms[@]}"
do
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	output_name=${PACKAGE_NAME}
	output_dir=$GOOS'-'$GOARCH
	awk -v line=$(tput cols) 'BEGIN{for (i = 1; i <= line; ++i){printf "-";}}'
	printf "\n\n"
	printf "${RED}DEBUG:${NORMAL} Now will be compiled for ${GOOS} - ${GOARCH} ....\n"

	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi
	env GOOS=$GOOS GOARCH=$GOARCH go build -o ${BIN_DIR}/${output_dir}/${output_name}
	if [ $? -ne 0 ]; then
		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
	if [ -e ${output_dir}/${PACKAGE_NAME}* ]
	then
		printf "${RED}DEBUG:${NORMAL} ${output_dir}/${output_name} ok\n"
	else
		printf "${RED}DEBUG:${NORMAL} ${output_dir}/${output_name} nok\n"
	fi
done



awk -v line=$(tput cols) 'BEGIN{for (i = 1; i <= line; ++i){printf "-";}}'
printf "\n\n"


printf "Zipping binaries....\n"
(cd ${BIN_DIR}; zip -r  ${PACKAGE_NAME}.zip .)
if [ -e ${BIN_DIR}/${PACKAGE_NAME}.zip ]
then
	printf "Zipped ok"
else
	printf "Zipped nok"
fi