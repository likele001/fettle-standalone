FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y gcc g++ && rm -rf /var/lib/apt/lists/*

COPY backend/ai-engine/requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY backend/ai-engine/ .

EXPOSE 20007 20008

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "20007"]
