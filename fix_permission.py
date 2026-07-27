import sys

f = "/www/wwwroot/fettle/backend/user-service/service/permission.go"
with open(f) as fh:
    c = fh.read()

# 1. Fix import - add uuid
c = c.replace(
    '"ai-platform/user-service/models"\n\n\t"gorm.io/gorm"',
    '"ai-platform/user-service/models"\n\n\t"github.com/google/uuid"\n\t"gorm.io/gorm"'
)

# 2. Fix TenantID assignment
c = c.replace('TenantID:    tenantID,', 'TenantID:    uuid.MustParse(tenantID),')

# 3. Fix RoleID/PermissionID assignment
c = c.replace(
    'RoleID:       roleID,\n\t\t\tPermissionID: permID,',
    'RoleID:       uuid.MustParse(roleID),\n\t\t\tPermissionID: uuid.MustParse(permID),'
)

with open(f, "w") as fh:
    fh.write(c)
print("fixed ok")
