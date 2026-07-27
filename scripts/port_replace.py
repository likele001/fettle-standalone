"""全局端口替换: 9xxx -> 2xxxx for fettle-standalone (v3)"""
import os, re, subprocess

WORKDIR = "/www/wwwroot/fettle-standalone"

PORT_MAP = {
    '9100': '20001',  '9200': '20002',  '9300': '20003',
    '9400': '20004',  '9500': '20005',  '9600': '20006',
    '9700': '20007',  '9701': '20008',  '9800': '20009',
    '9501': '20021',  '9601': '20022',  '9602': '20023',
    '9603': '20024',  '9604': '20025',
    '9432': '20010',  '9179': '20011',  '9177': '20012',
    '9422': '20013',  '9222': '20014',
    '9101': '20015',  '9102': '20016',
    '9530': '20017',  '9191': '20018',
    '9103': '20019',  '9104': '20020',
}

# 只处理这些前缀目录内且符合后缀的文件
INCLUDE_PATHS = [
    '.env', '.env.example',
    'scripts/start.sh',
    'deploy/docker/',
    'backend/gateway/', 'backend/user-service/', 'backend/agent-service/',
    'backend/chat-service/', 'backend/skill-service/', 'backend/billing-service/',
    'backend/ai-engine/',
    'frontend/web-admin/', 'frontend/web-platform/', 'frontend/web-website/',
    'plugins/',
    'database/',
]

def should_include(relpath):
    if relpath.startswith('backend/pkg/'):
        return False
    for p in INCLUDE_PATHS:
        if relpath.startswith(p):
            ext = os.path.splitext(relpath)[1]
            good_exts = {'.py', '.go', '.ts', '.vue', '.yml', '.yaml', '.sh', '.json', '.html', '.conf'}
            if ext in good_exts:
                return True
            basename = os.path.basename(relpath)
            if basename.startswith('.env') or basename == 'Dockerfile' or basename == 'nginx.conf':
                return True
    return False

def main():
    changed = 0
    for root, _, files in os.walk(WORKDIR):
        for fname in files:
            fpath = os.path.join(root, fname)
            rel = os.path.relpath(fpath, WORKDIR)
            if not should_include(rel):
                continue
            
            with open(fpath, 'r', errors='replace') as f:
                content = f.read()
            
            original = content
            for old_port, new_port in PORT_MAP.items():
                content = re.sub(rf'(?<!\d){old_port}(?!\d)', new_port, content)
            
            if content != original:
                with open(fpath, 'w') as f:
                    f.write(content)
                print(f"  ✓ {rel}")
                changed += 1
    
    print(f"\nChanged {changed} files")

if __name__ == '__main__':
    main()
