package cache

type CacheInterface interface {
	Get(shortCode string) (string, bool)
	Set(shortCode, longUrl string)
	Delete(shortCode string)
	Size() int
}
