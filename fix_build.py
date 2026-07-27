import sys

# Fix 1: agent-service handler/knowledge.go
f = "/www/wwwroot/fettle/backend/agent-service/handler/knowledge.go"
c = open(f).read()
c = c.replace(
    'KnowledgeBaseID: func() (id models.UUID) { id, _ = models.ParseUUID(kbID); return }(),',
    'KnowledgeBaseID: uuid.MustParse(kbID),'
)
c = c.replace(
    'if parsed, err := models.ParseUUID(tid); err == nil {',
    'if parsed, err := uuid.Parse(tid); err == nil {'
)
open(f, "w").write(c)
print("1. knowledge.go OK")

# Fix 2: skill-service router/router.go
f = "/www/wwwroot/fettle/backend/skill-service/router/router.go"
c = open(f).read()
c = c.replace('\t"gorm.io/gorm"\n', "")
open(f, "w").write(c)
print("2. skill router OK")

# Fix 3: billing-service service/billing_service.go
f = "/www/wwwroot/fettle/backend/billing-service/service/billing_service.go"
c = open(f).read()
c = c.replace(
    '\t// 获取套餐信息\n\tplan, err := s.repo.GetPlan(planID)\n\tif err != nil {\n\t\treturn err\n\t}',
    '\t// 获取套餐信息（验证套餐存在）\n\t_, err := s.repo.GetPlan(planID)\n\tif err != nil {\n\t\treturn err\n\t}'
)
open(f, "w").write(c)
print("3. billing OK")

print("All fixes applied")
