# SecretsRefresh Examples

## Описание

`SecretsRefresh` - это Custom Resource для автоматического отслеживания и обновления секретов в Kubernetes кластере.

## Возможности

- **Фильтрация по неймспейсам**: используйте `namespaceSelector` для выбора неймспейсов по меткам
- **Фильтрация по секретам**: используйте `secretSelector` для выбора конкретных секретов внутри отфильтрованных неймспейсов
- **Автоматическое отслеживание**: контроллер автоматически подписывается на изменения секретов в отфильтрованных неймспейсах

## Примеры использования

### 1. Простой пример (apps_v1alpha1_secretsrefresh_simple.yaml)

Отслеживание всех секретов с меткой `refresh: enabled` в неймспейсах с меткой `watch-secrets: true`:

```yaml
apiVersion: apps.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: secretsrefresh-simple
spec:
  namespaceSelector:
    matchLabels:
      watch-secrets: "true"
  secretSelector:
    matchLabels:
      refresh: "enabled"
  refreshInterval: "10m"
```

**Применение:**
```bash
kubectl apply -f apps_v1alpha1_secretsrefresh_simple.yaml
```

**Подготовка неймспейса:**
```bash
# Создать неймспейс с нужной меткой
kubectl create namespace my-namespace
kubectl label namespace my-namespace watch-secrets=true

# Создать секрет с нужной меткой
kubectl create secret generic my-secret \
  --from-literal=key=value \
  -n my-namespace
kubectl label secret my-secret refresh=enabled -n my-namespace
```

### 2. Расширенный пример (apps_v1alpha1_secretsrefresh.yaml)

Отслеживание секретов в production неймспейсах для команд backend и frontend:

```yaml
apiVersion: apps.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: secretsrefresh-sample
spec:
  namespaceSelector:
    matchLabels:
      environment: production
    matchExpressions:
      - key: team
        operator: In
        values:
          - backend
          - frontend
  secretSelector:
    matchLabels:
      auto-refresh: "true"
  refreshInterval: "5m"
```

**Применение:**
```bash
kubectl apply -f apps_v1alpha1_secretsrefresh.yaml
```

**Подготовка неймспейсов:**
```bash
# Backend production namespace
kubectl create namespace backend-prod
kubectl label namespace backend-prod environment=production team=backend

# Frontend production namespace
kubectl create namespace frontend-prod
kubectl label namespace frontend-prod environment=production team=frontend

# Создать секреты с auto-refresh
kubectl create secret generic db-credentials \
  --from-literal=password=secret \
  -n backend-prod
kubectl label secret db-credentials auto-refresh=true -n backend-prod
```

### 3. Отслеживание всех секретов во всех неймспейсах

Если не указать селекторы, будут отслеживаться все секреты во всех неймспейсах:

```yaml
apiVersion: apps.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: secretsrefresh-all
spec:
  refreshInterval: "15m"
```

## Проверка статуса

Проверить статус SecretsRefresh:

```bash
kubectl get secretsrefresh
kubectl describe secretsrefresh secretsrefresh-simple
```

Посмотреть логи контроллера:

```bash
kubectl logs -n traktor-system deployment/traktor-controller-manager -f
```

## Селекторы меток

### NamespaceSelector

Используется для фильтрации неймспейсов по меткам.

**Поддерживаемые операторы:**
- `In` - метка должна быть в списке значений
- `NotIn` - метка не должна быть в списке значений
- `Exists` - метка должна существовать
- `DoesNotExist` - метка не должна существовать

**Примеры:**

```yaml
# Только production
namespaceSelector:
  matchLabels:
    environment: production

# Production или staging
namespaceSelector:
  matchExpressions:
    - key: environment
      operator: In
      values:
        - production
        - staging

# Все, кроме development
namespaceSelector:
  matchExpressions:
    - key: environment
      operator: NotIn
      values:
        - development

# Все с меткой 'team'
namespaceSelector:
  matchExpressions:
    - key: team
      operator: Exists
```

### SecretSelector

Аналогично `namespaceSelector`, но применяется к секретам внутри отфильтрованных неймспейсов.

## RBAC Permissions

Контроллер требует следующие права:

```yaml
# Чтение SecretsRefresh CRD
- apiGroups: ["apps.gdxcloud.net"]
  resources: ["secretsrefreshes"]
  verbs: ["get", "list", "watch"]

# Обновление статуса
- apiGroups: ["apps.gdxcloud.net"]
  resources: ["secretsrefreshes/status"]
  verbs: ["get", "update", "patch"]

# Работа с секретами
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "list", "watch", "update", "patch"]

# Чтение неймспейсов
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list", "watch"]
```

## Troubleshooting

### Секреты не отслеживаются

1. Проверьте метки на неймспейсах:
   ```bash
   kubectl get namespaces --show-labels
   ```

2. Проверьте метки на секретах:
   ```bash
   kubectl get secrets -n <namespace> --show-labels
   ```

3. Проверьте условия в статусе:
   ```bash
   kubectl get secretsrefresh <name> -o jsonpath='{.status.conditions}' | jq
   ```

### Контроллер не запускается

Проверьте логи:
```bash
kubectl logs -n traktor-system deployment/traktor-controller-manager
```

Проверьте RBAC права:
```bash
kubectl get clusterrole manager-role -o yaml
```
