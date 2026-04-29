# Guía: Release 1.2 - Lo que resta (JWT + Login/Refresh/Logout + Middleware)

## ✅ Completado
- [x] Domain (user.go, session.go, role)
- [x] Ports (database.go)
- [x] Register use case + handler
- [x] PostgreSQL adapter (user_repo.go, session_repo.go)

---

## ⬜ Pendiente

### 1. Completar session_repo.go

Faltan `FindByRefreshToken` y `Revoke`:

```go
func (r *SessionRepository) FindByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
    var s domain.Session
    err := r.db.WithContext(ctx).Where("refresh_token = ?", token).First(&s).Error
    return &s, err
}

func (r *SessionRepository) Revoke(ctx context.Context, token string) error {
    return r.db.WithContext(ctx).Model(&domain.Session{}).
        Where("refresh_token = ?", token).
        Update("revoked_at", time.Now()).Error
}
```

---

### 2. Adapter JWT (`internal/infrastructure/auth/jwt.go`)

Instala: `go get github.com/golang-jwt/jwt/v5`

Funciones a implementar:

**Claims struct:**
```go
type Claims struct {
    UserID int    `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}
```

**GenerateAccessToken** - JWT con expiración configurable (15min por defecto), firmado con HMAC.

**GenerateRefreshToken** - 32 bytes aleatorios con `crypto/rand`, retorna como hex string.

**ValidateAccessToken** - ParseWithClaims, validar expiración, retornar Claims si es válido.

---

### 3. Login en auth_usecase.go

```go
func (au *AuthUsecase) Login(ctx context.Context, email, password string) (*LoginResult, error) {
    // 1. FindByEmail
    // 2. bcrypt.CompareHashAndPassword
    // 3. GenerateAccessToken(userID, role, secret)
    // 4. GenerateRefreshToken
    // 5. SessionRepository.Create(accessToken, refreshToken, expiresAt)
    // 6. Return LoginResult{AccessToken, RefreshToken}
}
```

**Refresh:** FindByRefreshToken → validar no expirado/revoked → generar nuevo access token → actualizar en DB.

**Logout:** Revoke(refreshToken).

---

### 4. Handlers Login/Refresh/Logout

En `auth_handler.go` agregar:

| Endpoint | Método | Body | Response |
|----------|--------|------|----------|
| POST /api/v1/auth/login | Login | email, password | access_token, refresh_token |
| POST /api/v1/auth/refresh | Refresh | refresh_token | access_token, refresh_token |
| POST /api/v1/auth/logout | Logout | refresh_token | mensaje |

Cada uno sigue el patrón: bind JSON → llamar use case → responder JSON o error.

---

### 5. JWT Middleware

Archivo: `internal/infrastructure/rest/middleware/jwt.go`

```go
func AuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extraer header: "Authorization: Bearer <token>"
        // 2. Validar token con ValidateAccessToken
        // 3. Si inválido: c.AbortWithStatusJSON(401)
        // 4. Si válido: c.Set("user_id", claims.UserID), c.Set("role", claims.Role)
        // 5. c.Next()
    }
}
```

Protege endpoints en el router:

```go
router.POST("/api/v1/ask", AuthMiddleware(cfg.JWTSecret), handler.Ask)
router.POST("/api/v1/ingest", AuthMiddleware(cfg.JWTSecret), handler.Ingest)
```

---

### 6. Agregar rutas en handlers.go

```go
type Handlers struct {
    AskHandler    *AskHandler
    IngestHandler *IngestHandler
    AuthHandler   *AuthHandler
}
```

Registrar en `NewRouter`:
```go
router.POST("/api/v1/auth/register", h.AuthHandler.Register)
router.POST("/api/v1/auth/login", h.AuthHandler.Login)
router.POST("/api/v1/auth/refresh", h.AuthHandler.Refresh)
router.POST("/api/v1/auth/logout", h.AuthHandler.Logout)
```

---

### 7. Config

Agregar al struct y .env:
- `DATABASE_URL` (ya tienes)
- `JWT_SECRET`
- `JWT_ACCESS_EXPIRATION` (default: "15m")
- `JWT_REFRESH_EXPIRATION` (default: "720h")

---

### Orden sugerido

```
1. session_repo.go → FindByRefreshToken, Revoke
2. jwt.go → GenerateAccessToken, GenerateRefreshToken, ValidateAccessToken
3. auth_usecase.go → Login, Refresh, Logout
4. auth_handler.go → Login, Refresh, Logout handlers
5. jwt middleware
6. handlers.go → rutas faltantes
7. main.go → integrar todo
```
