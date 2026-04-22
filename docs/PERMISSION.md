# Добавление нового permission

## 1. Объявить permission


Файл: [`/internal/model/authz/permission.go`](../internal/model/authz/permission.go)

```go
const PermMenuCreate Permission = "menu:create"
```

Правило описания:
```
<ресурс>:<действие>
```

Примеры:
```
menu:create
order:read
table:release
```

## 2. Добавить в policy

Файл: [`/internal/model/authz/policy.go`](../internal/model/authz/permission.go)

```go
var rolePermissions = map[UserRole]permissionSet{
	RoleAdmin: {
		PermUpdatePassword: {}, // Новое разрешение для роли Admin
	},
	RoleManager: {
		PermUpdatePassword: {},
	},
	RoleWaiter: {
		PermUpdatePassword: {},
	},
}
```

## 3. Добавить middleware для проверки прав

Файл: [`/cmd/service/main.go`](../internal/model/authz/permission.go)

```go
r.POST("/menu",
    middleware.RequirePermission(authz.PermMenuCreate),
    handler.CreateMenu,
)
```

## 4. Не проверять роли вручную

НЕ проверять права `вручную`:
```go
if role == RoleAdmin {
    ...
}
```
Использовать `middleware`:
```go
middleware.RequirePermission(PermMenuCreate)
```

## 5. Результат

1. нет токена -> 401 Unauthorized
2. нет permission -> 403 Forbidden
3. есть permission -> доступ разрешён
