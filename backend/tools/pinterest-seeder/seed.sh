export KARTONKO_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3OTE1NzQ4MDAsInVzZXJJZCI6MX0.dF7y_U8efkdAjWyc8K1xra12KEiIkAyGC2LE9ZQMtOw"
python3 seed.py \
    --endpoint http://localhost:3000/image/upload/batch \
    --batch-size 30 \
    --pinterest-limit 100 \
    --limit 10000 \
    --concurrency 16 \
    --tag pinterest \
    --query "lulz"
    