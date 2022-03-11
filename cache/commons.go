package cache

type CacheProvider interface {
	Get(k string) (data interface{}, found bool)
	Delete(k string)
	Set(k string, data interface{})
	SpawnJanitorGoroutine()
}
