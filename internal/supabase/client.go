package supabase

import (
	"os"
	"sync"

	supa "github.com/supabase-community/supabase-go"
)

var (
	client *supa.Client
	once   sync.Once
)

// GetClient mengembalikan singleton Supabase client (aman untuk concurrent use)
func GetClient() *supa.Client {
	once.Do(func() {
		url := os.Getenv("SUPABASE_URL")
		key := os.Getenv("SUPABASE_KEY") // harus service_role key untuk backend full access

		if url == "" || key == "" {
			panic("SUPABASE_URL atau SUPABASE_KEY tidak ditemukan di environment")
		}

		client, _ = supa.NewClient(url, key, nil)
		// Note: error diabaikan di sini untuk simplicity, tapi di production bisa log
	})

	return client
}