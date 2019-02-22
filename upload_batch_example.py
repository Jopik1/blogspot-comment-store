import aiohttp
import asyncio
from aiohttp import FormData

async def upload(batch_id,batch_key,worker_id,version, filename):
    async with aiohttp.ClientSession() as session:
        url = 'http://localhost/submitBatchWorkUnit'
        data = FormData()

        data.add_field('batchID',batch_id)
        data.add_field('batchKey', batch_key)
        data.add_field('workerID', worker_id)
        data.add_field('version', version)
        data.add_field('data',
                       open(filename, 'rb'),
                       filename=filename,
                       content_type='application/x-gzip')

        resp=await session.post(url, data=data)
        print(resp.status)
        print(await resp.text())


loop = asyncio.get_event_loop()
# Blocking call which returns when the display_date() coroutine is done
loop.run_until_complete(upload(batch_id='11111',batch_key='222222',worker_id='333333',version='1',filename='test.json.gz'))
loop.close()

