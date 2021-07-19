package internal

type CompanyStats struct {
	OnlineUsers uint         `json:"online_users"`
	MaxUsers    uint         `json:"max_users"`
	Users       []ClientStat `json:"active_users"`
}
type ClientStat struct {
	UserId string `json:"user_id"`
}

type Cache struct {
	Info *CompanyStats
}

func NewCache() *Cache {
	return &Cache{
		Info: new(CompanyStats),
	}
}
func (c *Cache) update(stats *CompanyStats) {
	c.Info = stats
}