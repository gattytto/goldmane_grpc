from tigera.goldmane_v1.services import flows,statistics
from tigera.goldmane_v1.services.flows.transports.grpc_asyncio import FlowsGrpcAsyncIOTransport
from tigera.goldmane_v1.services.statistics.transports.grpc_asyncio import StatisticsGrpcAsyncIOTransport
from grpc.experimental import aio
from grpc import ssl_channel_credentials
from google.api_core import client_options
from google.auth.credentials import AnonymousCredentials
from tigera.goldmane_v1.types.flows_service import FlowStreamRequest,FlowResult
from tigera.goldmane_v1.types.statistics_service import StatisticsRequest,StatisticsResult
from typing import AsyncIterable
from google.protobuf.json_format import MessageToJson
import asyncio,json
from elasticsearch import Elasticsearch
import urllib3
import logging

# Disable the InsecureRequestWarning specifically
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# For warnings related to the underlying transport layer (e.g., product verification)
logging.getLogger('elastic_transport.transport').setLevel(logging.ERROR)

# For general Elasticsearch client warnings
logging.getLogger('elasticsearch').setLevel(logging.ERROR)

# Alternatively, to completely suppress all warnings from these loggers:
# logging.getLogger('elastic_transport.transport').setLevel(logging.CRITICAL)
# logging.getLogger('elasticsearch').setLevel(logging.CRITICAL)
import os
es = Elasticsearch(
    hosts=["https://elasticsearch-sample-es-http.elastic-system.svc:9200"],  # Update with your ES host
    basic_auth=("elastic", os.getenv('ES_PASSWORD')),  # Update with your user and password
    verify_certs=False,
    ssl_show_warn=False
)

GOLDMANE_SERVICE="goldmane.calico-system.svc"

kk=open(os.getenv(SSL_KEY_PATH)).read()
bd=open(os.getenv(SSL_CERT_PATH)).read()
creds = ssl_channel_credentials(bd,kk,bd)

async def statistics_stream():
    client=statistics.StatisticsAsyncClient(
        client_options=client_options.ClientOptions(
            api_endpoint=GOLDMANE_SERVICE+":7443"
        ),
        transport=StatisticsGrpcAsyncIOTransport(
            host=GOLDMANE_SERVICE+":7443",
            credentials=AnonymousCredentials(),
            channel=aio.secure_channel(
                target=GOLDMANE_SERVICE+":7443",
                credentials=creds
                )
        )
    )

    things:AsyncIterable[StatisticsResult]
    try:
        things= await client.list(request=StatisticsRequest())
    except Exception as e:
        print(f'ERROR {e}')
        return
    async for resp in things:
        d=json.loads(MessageToJson(resp._pb))
        try:
            response = es.index(index="goldmane.statistics", body=d)
            print(response)
        except Exception as e:
            print(f'ERROR {e}')

async def flows_stream():
    global es
    client=flows.FlowsAsyncClient(
        client_options=client_options.ClientOptions(
            api_endpoint=GOLDMANE_SERVICE+":7443"
        ),
        transport=FlowsGrpcAsyncIOTransport(
            host=GOLDMANE_SERVICE+":7443",
            credentials=AnonymousCredentials(),
            channel=aio.secure_channel(
                target=GOLDMANE_SERVICE+":7443",
                credentials=creds
                )
        )
    )
    
    things:AsyncIterable[FlowResult]
    try:
        things= await client.stream(request=FlowStreamRequest())
    except Exception as e:
        print(f'ERROR {e}')
        return
    async for resp in things:
        d=json.loads(MessageToJson(resp.flow._pb))
        try:
            response = es.index(index="goldmane.flows", body=d)
            print(response)
        except Exception as e:
            print(f'ERROR {e}')

if __name__ == '__main__':
    print('in main')
    
    loop=asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    try:
        tsk=loop.create_task(statistics_stream())
        tsk2=loop.create_task(flows_stream())
        loop.run_forever()
    except KeyboardInterrupt:
        tsk.cancel()
        tsk2.cancel()
        print('we are done')
        loop.stop()
    except Exception as e:
        loop.stop()
        print(f'error {e}')
    finally:
        print("bye")


