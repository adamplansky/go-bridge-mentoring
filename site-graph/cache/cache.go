package cache

type Cache interface {
	// Add insert new key-value into cache, if key already exists
	// in cache returns evicted=true
	Add(key string, value interface{})
	// Get gets value from the coresponding key from cache
	// ok == true if object exist in cache, otherwire ok == false
	Get(key string) (value interface{}, ok bool)
}
