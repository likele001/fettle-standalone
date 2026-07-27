#!/bin/bash
set -e

# MinIO 原生安装
cat <<'EOF'
--- Installing MinIO ---
EOF
sudo mkdir -p /usr/local/bin
sudo curl -L https://dl.min.io/server/minio/release/linux-amd64/minio -o /usr/local/bin/minio
sudo chmod +x /usr/local/bin/minio
sudo mkdir -p /var/minio/data
sudo chown -R www:www /var/minio/data || true
cat <<'EOF' | sudo tee /etc/systemd/system/minio.service
[Unit]
Description=MinIO
After=network.target

[Service]
User=www
Group=www
ExecStart=/usr/local/bin/minio server /var/minio/data --address :9000
Restart=always
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl enable --now minio

# NATS 原生安装
cat <<'EOF'
--- Installing NATS ---
EOF
sudo curl -L -f https://github.com/nats-io/nats-server/releases/download/v2.10.9/nats-server-v2.10.9-linux-amd64.tar.gz -o /tmp/nats-server.tar.gz
sudo tar xzf /tmp/nats-server.tar.gz -C /tmp
sudo mv /tmp/nats-server-v2.10.9-linux-amd64/nats-server /usr/local/bin/
sudo chmod +x /usr/local/bin/nats-server
cat <<'EOF' | sudo tee /etc/systemd/system/nats.service
[Unit]
Description=NATS Server
After=network.target

[Service]
ExecStart=/usr/local/bin/nats-server -a 0.0.0.0 -p 4222
Restart=always
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
sudo systemctl daemon-reload
sudo systemctl enable --now nats

# Milvus 原生安装（简化示例）
cat <<'EOF'
--- Installing Milvus ---
EOF
sudo mkdir -p /usr/local/milvus
sudo curl -L -f https://github.com/milvus-io/milvus/releases/download/v2.3.0/milvus-standalone-linux-amd64.tar.gz -o /tmp/milvus.tar.gz
sudo tar xzf /tmp/milvus.tar.gz -C /tmp
sudo mv /tmp/milvus/* /usr/local/milvus/
sudo mkdir -p /var/lib/milvus
sudo chown -R www:www /var/lib/milvus || true
# 由于 Milvus 复杂依赖，建议按官方说明配置并启用
cat <<'EOF'
Milvus binary installed under /usr/local/milvus
请根据官方文档完成环境设置和启动。
EOF