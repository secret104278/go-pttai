#!/bin/bash

branch=`git branch|grep '*'|awk '{print $2}'`

p2pport=9487
port=9774
apiport=14779
theexthttpaddr=""
exthttpaddr="http://localhost:${port}"
theextrpcaddr=""
extrpcaddr="http://localhost:${apiport}"

while [[ $# -gt 0 ]]
do
    key="$1"
    case $key in
        -p|--port)
        port="$2"
        exthttpaddr="http://localhost:${port}"
        shift # past argument
        shift # past value
        ;;
        -a|--apiport)
        apiport="$2"
        extrpcaddr="http://localhost:${apiport}"
        shift # past argument
        shift # past value
        ;;
        -q|--p2pport)
        p2pport="$2"
        shift # past argument
        shift # past value
        ;;
        -e|--exthttpaddr)
        theexthttpaddr="$2"
        shift # past argument
        shift # past value
        ;;
        -b|--extrpcaddr)
        theextrpcaddr="$2"
        shift # past argument
        shift # past value
        ;;
    esac
done

if [ "${theexthttpaddr}" != "" ]
then
    exthttpaddr=${theexthttpaddr}
fi

if [ "${theextrpcaddr}" != "" ]
then
    extrpcaddr=${theextrpcaddr}
fi

echo ""
echo "${exthttpaddr}"
echo ""
echo "${extrpcaddr}"
echo ""
sleep 1

docker run --rm -itd -p ${p2pport}:9487 -p ${port}:9774 -p ${apiport}:14779 --name go-pttai_${branch} go-pttai:${branch} gptt "--testp2p" "--httpdir" "/static" "--httpaddr" "0.0.0.0:9774" "--rpcaddr" "0.0.0.0" --exthttpaddr "${exthttpaddr}" --extrpcaddr "${extrpcaddr}"
