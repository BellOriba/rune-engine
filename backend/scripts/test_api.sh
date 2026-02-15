#!/bin/bash

URL="http://localhost:8080/v1/convert?width=160"
IMAGE_PATH="example.png"
TOTAL_REQUESTS=12

echo "--- Iniciando Teste de Rate Limit e Cache ---"

for i in $(seq 1 $TOTAL_REQUESTS); do
    echo -n "Requisição #$i: "
    
    RESPONSE=$(curl -s -i -X POST "$URL" \
        -F "image=@$IMAGE_PATH" \
        -H "Content-Type: multipart/form-data")

    STATUS=$(echo "$RESPONSE" | grep "HTTP/" | awk '{print $2}')
    CACHE=$(echo "$RESPONSE" | grep -i "X-Cache:" | awk '{print $2}' | tr -d '\r')
    
    if [ "$STATUS" == "429" ]; then
        echo -e "\e[31mBloqueado (429 Too Many Requests)\e[0m"
    elif [ "$STATUS" == "200" ]; then
        echo -ne "\e[32mSucesso (200)\e[0m"
        if [ ! -z "$CACHE" ]; then
            echo -e " | Cache: \e[34m$CACHE\e[0m"
        else
            echo -e " | Cache: \e[33mMISS (ou não enviado)\e[0m"
        fi
    else
        echo "Status: $STATUS"
    fi
done

echo "--- Teste Finalizado ---"

