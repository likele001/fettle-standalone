#!/bin/bash
# 以宝塔面板为唯一权威重启 standalone 服务：先换二进制，再调面板 stop/start，最后按端口+健康校验。
# 用法: restart-standalone.sh <面板项目名> <二进制名> <端口>
set -u
NAME=$1; BIN=$2; PORT=$3
PY=/www/server/panel/pyenv/bin/python3
BUILD=/www/wwwroot/fettle-standalone/build
NEW=/tmp/sabuild/$BIN
LOG=/var/log/fettle/panel-restart.log
ts() { date '+%F %T'; }

panel() {  # $1=stop|start
  (cd /www/server/panel && $PY -c "
import sys, os
sys.path.insert(0, '/www/server/panel/class'); sys.path.insert(0, '/www/server/panel')
from projectModel import goModel
import public
get = public.dict_obj(); get.project_name = '$NAME'
print(goModel.main().${1}_project(get))
" 2>&1)
}

echo "$(ts) ==== $NAME ($BIN :$PORT) ====" | tee -a "$LOG"
install -o fettle-std -g fettle-std -m 755 "$NEW" "$BUILD/$BIN" || { echo "INSTALL FAILED"; exit 1; }
echo "$(ts) swapped binary" | tee -a "$LOG"
echo "$(ts) stop -> $(panel stop)" | tee -a "$LOG"
for i in $(seq 1 10); do
  ss -tlnH 2>/dev/null | grep -q ":$PORT\b" || break
  sleep 1
done
echo "$(ts) start -> $(panel start)" | tee -a "$LOG"
for i in $(seq 1 20); do
  code=$(curl -s -m 3 -o /dev/null -w '%{http_code}' "http://127.0.0.1:$PORT/health" 2>/dev/null)
  [ "$code" = "200" ] && { echo "$(ts) OK $NAME pid=$(cat /var/tmp/gopids/$NAME.pid 2>/dev/null) health=$code" | tee -a "$LOG"; exit 0; }
  sleep 1
done
echo "$(ts) FAIL $NAME 端口 $PORT 健康检查未通过 (last=$code)" | tee -a "$LOG"
exit 1
