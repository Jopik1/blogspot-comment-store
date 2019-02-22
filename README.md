
End points: 

http://blogstore.bot.nu/submitBatchUnit

POST

Submits a Batch Unit for storage via Multipart Post,

expects the following parameters:

batchID - Batch ID

batchKey - Batch Random Key (supplied by the master)

workerID - Worker ID (supplied by the master)

version - The version of the Worker (same as json format version)

data - The file data which is stored (json.gz) 

The file is stored in a directory tree when the first digit of the 

batchID is the directory name and the filename is batchID.batchKey.json.gz

For example for batchID=15 and batchKey=983 the file will be stored 

as /1/15.983.json.gz

If the file with the exact name already exists, it will be overwritten.


Python example in: upload_batch_example.py


http://blogstore.bot.nu/getVerifyBatchUnit?batchID=11111&batchKey=222222

GET

Verifies that Batch Unit exists, returns JSON with size of file

Returns 404 if Batch Unit does not exist.


http://blogstore.bot.nu/getBatchUnit?batchID=11111&batchKey=222222

GET

Retrieves a single batch unit (downloads the file)

Returns 404 if Batch Unit does not exist.


http://blogstore.bot.nu/uploadedBatches

GET

Lists directories with all uploaded batches and allows download
