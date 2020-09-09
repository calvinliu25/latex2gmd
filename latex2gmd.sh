#!/bin/bash
# uncomment the line bellow of server.py does not terminate itself
# but this should not be necessary in the final release
# trap "kill 0" EXIT

if [ $# -eq 1 ]
  then
    echo "Missing arguments for input file or output file"
    echo "Please run: ./latex2gmd.sh <inputfile> <outputfile>"
    echo "example: ./latex2gmd.sh sample.tex output.md"
    exit 1
fi

latexFile=$1 # file name for input
mdFile=$2    # file name for output
echo "Latex -> GMD: $latexFile -> $mdFile"

python3 server.py $mdFile &
go run client.go $latexFile
