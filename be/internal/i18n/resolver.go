package i18n

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"falzo-be/pkg/cache"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	DefaultLocale = "vi"
	English       = "en"
	Vietnamese    = "vi"

	localeHeader       = "X-Locale"
	redisFailureLimit  = 3
	redisCooldown      = 30 * time.Second
	redisCallTimeout   = 100 * time.Millisecond
	dbCallTimeout      = 500 * time.Millisecond
	postgresRetryDelay = 2 * time.Second
	reloadDebounce     = 300 * time.Millisecond
)

type Translation struct {
	Key    string
	Locale string
	Value  string
}

type Resolver struct {
	db    *pgxpool.Pool
	redis *goredis.Client

	mu       sync.RWMutex
	messages map[string]map[string]string

	redisMu            sync.Mutex
	redisFailures      int
	redisDisabledUntil time.Time
}

var defaultResolverMu sync.RWMutex
var defaultResolver *Resolver

func NewResolver(db *pgxpool.Pool, redisClient cache.Client) *Resolver {
	resolver := &Resolver{
		db:       db,
		messages: builtInMessages(),
	}

	if redisClient != nil {
		resolver.redis = redisClient.Client()
	}

	return resolver
}

func SetDefaultResolver(resolver *Resolver) {
	defaultResolverMu.Lock()
	defer defaultResolverMu.Unlock()
	defaultResolver = resolver
}

func ResolveRequest(r *http.Request, key string) Translation {
	locale := LocaleFromRequest(r)

	defaultResolverMu.RLock()
	resolver := defaultResolver
	defaultResolverMu.RUnlock()

	if resolver == nil {
		normalizedKey := normalizeKey(key)
		return Translation{Key: normalizedKey, Locale: locale, Value: normalizedKey}
	}

	ctx := context.Background()
	if r != nil {
		ctx = r.Context()
	}

	return resolver.Resolve(ctx, key, locale)
}

func LocaleFromRequest(r *http.Request) string {
	if r == nil {
		return DefaultLocale
	}

	if locale := normalizeLocale(r.Header.Get(localeHeader)); locale != "" {
		return locale
	}

	for _, part := range strings.Split(r.Header.Get("Accept-Language"), ",") {
		value := strings.TrimSpace(strings.Split(part, ";")[0])
		if locale := normalizeLocale(value); locale != "" {
			return locale
		}
	}

	return DefaultLocale
}

func (r *Resolver) Load(ctx context.Context) error {
	if r == nil || r.db == nil {
		return nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, err := r.db.Query(queryCtx, `
		SELECT message_key, locale, value
		FROM i18n_messages
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	messages := builtInMessages()
	for rows.Next() {
		var key, locale, value string
		if err := rows.Scan(&key, &locale, &value); err != nil {
			return err
		}
		putMessage(messages, locale, key, value)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.messages = messages
	r.mu.Unlock()

	r.writeRedisSnapshot(ctx, messages)
	return nil
}

func (r *Resolver) Resolve(ctx context.Context, key string, locale string) Translation {
	normalizedKey := normalizeKey(key)
	normalizedLocale := normalizeLocale(locale)
	if normalizedLocale == "" {
		normalizedLocale = DefaultLocale
	}
	if normalizedKey == "" {
		return Translation{Key: "", Locale: normalizedLocale, Value: ""}
	}

	if value, ok := r.lookupMemory(normalizedLocale, normalizedKey); ok {
		return Translation{Key: normalizedKey, Locale: normalizedLocale, Value: value}
	}
	if normalizedLocale != English {
		if value, ok := r.lookupMemory(English, normalizedKey); ok {
			return Translation{Key: normalizedKey, Locale: normalizedLocale, Value: value}
		}
	}

	if value, ok := r.lookupRedis(ctx, normalizedLocale, normalizedKey); ok {
		r.setMemory(normalizedLocale, normalizedKey, value)
		return Translation{Key: normalizedKey, Locale: normalizedLocale, Value: value}
	}
	if normalizedLocale != English {
		if value, ok := r.lookupRedis(ctx, English, normalizedKey); ok {
			r.setMemory(English, normalizedKey, value)
			return Translation{Key: normalizedKey, Locale: normalizedLocale, Value: value}
		}
	}

	if value, ok := r.lookupDB(ctx, normalizedLocale, normalizedKey); ok {
		r.setMemory(normalizedLocale, normalizedKey, value)
		r.setRedis(ctx, normalizedLocale, normalizedKey, value)
		return Translation{Key: normalizedKey, Locale: normalizedLocale, Value: value}
	}
	if normalizedLocale != English {
		if value, ok := r.lookupDB(ctx, English, normalizedKey); ok {
			r.setMemory(English, normalizedKey, value)
			r.setRedis(ctx, English, normalizedKey, value)
			return Translation{Key: normalizedKey, Locale: normalizedLocale, Value: value}
		}
	}

	return Translation{Key: normalizedKey, Locale: normalizedLocale, Value: normalizedKey}
}

func (r *Resolver) RunPostgresListener(ctx context.Context) {
	if r == nil || r.db == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := r.db.Acquire(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("i18n listener could not acquire postgres connection")
			sleepOrDone(ctx, postgresRetryDelay)
			continue
		}

		if _, err := conn.Exec(ctx, "LISTEN i18n_messages_changed"); err != nil {
			conn.Release()
			log.Warn().Err(err).Msg("i18n listener could not subscribe")
			sleepOrDone(ctx, postgresRetryDelay)
			continue
		}

		log.Info().Msg("i18n listener subscribed")
		for {
			if err := conn.Conn().PgConn().WaitForNotification(ctx); err != nil {
				conn.Release()
				if !errors.Is(err, context.Canceled) {
					log.Warn().Err(err).Msg("i18n listener disconnected")
				}
				break
			}

			sleepOrDone(ctx, reloadDebounce)
			if err := r.Load(ctx); err != nil {
				log.Warn().Err(err).Msg("i18n cache reload failed")
			} else {
				log.Info().Msg("i18n cache reloaded")
			}
		}
	}
}

func (r *Resolver) lookupMemory(locale, key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	value, ok := r.messages[locale][key]
	return value, ok
}

func (r *Resolver) setMemory(locale, key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	putMessage(r.messages, locale, key, value)
}

func (r *Resolver) lookupDB(ctx context.Context, locale, key string) (string, bool) {
	if r == nil || r.db == nil {
		return "", false
	}

	queryCtx, cancel := context.WithTimeout(ctx, dbCallTimeout)
	defer cancel()

	var value string
	err := r.db.QueryRow(queryCtx, `
		SELECT value
		FROM i18n_messages
		WHERE message_key = $1 AND locale = $2
	`, key, locale).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false
	}
	if err != nil {
		log.Warn().Err(err).Str("locale", locale).Str("message_key", key).Msg("i18n db lookup failed")
		return "", false
	}

	return value, true
}

func (r *Resolver) lookupRedis(ctx context.Context, locale, key string) (string, bool) {
	if r == nil || r.redis == nil || r.shouldSkipRedis() {
		return "", false
	}

	redisCtx, cancel := context.WithTimeout(ctx, redisCallTimeout)
	defer cancel()

	value, err := r.redis.HGet(redisCtx, redisHashKey(locale), key).Result()
	if errors.Is(err, goredis.Nil) {
		return "", false
	}
	if err != nil {
		r.markRedisFailure()
		return "", false
	}

	r.markRedisSuccess()
	return value, true
}

func (r *Resolver) setRedis(ctx context.Context, locale, key, value string) {
	if r == nil || r.redis == nil || r.shouldSkipRedis() {
		return
	}

	redisCtx, cancel := context.WithTimeout(ctx, redisCallTimeout)
	defer cancel()

	if err := r.redis.HSet(redisCtx, redisHashKey(locale), key, value).Err(); err != nil {
		r.markRedisFailure()
		return
	}
	r.markRedisSuccess()
}

func (r *Resolver) writeRedisSnapshot(ctx context.Context, messages map[string]map[string]string) {
	if r == nil || r.redis == nil || r.shouldSkipRedis() {
		return
	}

	redisCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	pipe := r.redis.Pipeline()
	for locale, localeMessages := range messages {
		hashKey := redisHashKey(locale)
		pipe.Del(redisCtx, hashKey)
		if len(localeMessages) > 0 {
			values := make(map[string]any, len(localeMessages))
			for key, value := range localeMessages {
				values[key] = value
			}
			pipe.HSet(redisCtx, hashKey, values)
		}
	}

	if _, err := pipe.Exec(redisCtx); err != nil {
		r.markRedisFailure()
		return
	}
	r.markRedisSuccess()
}

func (r *Resolver) shouldSkipRedis() bool {
	r.redisMu.Lock()
	defer r.redisMu.Unlock()
	return time.Now().Before(r.redisDisabledUntil)
}

func (r *Resolver) markRedisFailure() {
	r.redisMu.Lock()
	defer r.redisMu.Unlock()
	r.redisFailures++
	if r.redisFailures >= redisFailureLimit {
		r.redisDisabledUntil = time.Now().Add(redisCooldown)
	}
}

func (r *Resolver) markRedisSuccess() {
	r.redisMu.Lock()
	defer r.redisMu.Unlock()
	r.redisFailures = 0
	r.redisDisabledUntil = time.Time{}
}

func normalizeLocale(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(value, English) {
		return English
	}
	if strings.HasPrefix(value, Vietnamese) {
		return Vietnamese
	}
	return ""
}

func normalizeKey(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func redisHashKey(locale string) string {
	return "i18n:messages:" + locale
}

func putMessage(messages map[string]map[string]string, locale, key, value string) {
	locale = normalizeLocale(locale)
	key = normalizeKey(key)
	if locale == "" || key == "" {
		return
	}
	if _, ok := messages[locale]; !ok {
		messages[locale] = map[string]string{}
	}
	messages[locale][key] = value
}

func sleepOrDone(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func builtInMessages() map[string]map[string]string {
	return map[string]map[string]string{
		English: {
			"Validation field":      "Validation field",
			"Invalid credentials":   "Invalid credentials",
			"Internal server error": "Internal server error",
			"Unexpected error":      "Unexpected error",
			"Unauthorized":          "Unauthorized",
			"Forbidden":             "Forbidden",
			"Too many requests":     "Too many requests",
		},
		Vietnamese: {
			"Validation field":      "Dữ liệu không hợp lệ",
			"Invalid credentials":   "Thông tin đăng nhập không hợp lệ",
			"Internal server error": "Lỗi máy chủ",
			"Unexpected error":      "Lỗi không mong muốn",
			"Unauthorized":          "Chưa xác thực",
			"Forbidden":             "Không có quyền truy cập",
			"Too many requests":     "Quá nhiều yêu cầu",
		},
	}
}
